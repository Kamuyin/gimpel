package modules

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent/config"
	"gimpel/internal/agent/module"
	"gimpel/internal/agent/store"
)

type ListenerStarter interface {
	StartListener(ctx context.Context, cfg config.ListenerConfig) error
	StopListener(id string) error
}

type Reconciler struct {
	store           Store
	downloader      *ModuleDownloader
	supervisor      *module.Supervisor
	listenerStarter ListenerStarter
}

func NewReconciler(store Store, downloader *ModuleDownloader, supervisor *module.Supervisor) *Reconciler {
	return &Reconciler{
		store:      store,
		downloader: downloader,
		supervisor: supervisor,
	}
}

func (r *Reconciler) SetListenerStarter(ls ListenerStarter) {
	r.listenerStarter = ls
}

func (r *Reconciler) Reconcile(ctx context.Context) error {
	deployment, err := r.store.GetDeploymentConfig()
	if err != nil {
		return fmt.Errorf("getting deployment config: %w", err)
	}

	if deployment == nil {
		log.Debug("no deployment config yet")
		return nil
	}

	log.WithField("modules", len(deployment.Modules)).Info("reconciling module deployments")

	running := r.supervisor.ListModules()
	runningMap := make(map[string]bool)
	for _, info := range running {
		runningMap[info.ID] = true
	}

	for _, modDeploy := range deployment.Modules {
		if !modDeploy.Enabled {
			continue
		}

		moduleKey := fmt.Sprintf("%s:%s", modDeploy.ModuleID, modDeploy.ModuleVersion)

		if runningMap[modDeploy.ModuleID] {
			log.WithField("module", modDeploy.ModuleID).Debug("module already running")
			delete(runningMap, modDeploy.ModuleID)
			continue
		}

		cached, err := r.downloader.DownloadModule(ctx, modDeploy.ModuleID, modDeploy.ModuleVersion)
		if err != nil {
			log.WithError(err).WithField("module", moduleKey).Error("failed to download module")
			continue
		}

		modCfg := r.deploymentToConfig(modDeploy, cached)

		if err := r.supervisor.StartModule(ctx, modCfg); err != nil {
			log.WithError(err).WithField("module", modDeploy.ModuleID).Error("failed to start module")
			continue
		}

		if r.listenerStarter != nil {
			for _, lCfg := range modCfg.Listeners {
				if err := r.listenerStarter.StartListener(ctx, lCfg); err != nil {
					log.WithError(err).WithFields(log.Fields{
						"module":   modDeploy.ModuleID,
						"listener": lCfg.ID,
						"port":     lCfg.Port,
					}).Error("failed to start listener")
				} else {
					log.WithFields(log.Fields{
						"module":   modDeploy.ModuleID,
						"listener": lCfg.ID,
						"port":     lCfg.Port,
					}).Info("listener started")
				}
			}
		}

		log.WithFields(log.Fields{
			"module":  modDeploy.ModuleID,
			"version": modDeploy.ModuleVersion,
		}).Info("module started")
	}

	for moduleID := range runningMap {
		log.WithField("module", moduleID).Info("stopping unassigned module")
		if err := r.supervisor.StopModule(ctx, moduleID); err != nil {
			log.WithError(err).WithField("module", moduleID).Warn("failed to stop module")
		}
	}

	return nil
}

func (r *Reconciler) deploymentToConfig(deploy store.ModuleDeployment, cache *store.ModuleCache) config.ModuleConfig {
	listeners := make([]config.ListenerConfig, 0, len(deploy.Listeners))
	for _, l := range deploy.Listeners {
		listeners = append(listeners, config.ListenerConfig{
			ID:              l.ID,
			Protocol:        l.Protocol,
			Port:            int(l.Port),
			ModuleID:        deploy.ModuleID,
			HighInteraction: l.HighInteraction,
		})
	}

	execMode := deploy.ExecutionMode
	if execMode == "" {
		execMode = "containerd"
	}

	return config.ModuleConfig{
		ID:            deploy.ModuleID,
		Name:          deploy.ModuleID,
		Image:         cache.ImagePath,
		Env:           deploy.Env,
		Listeners:     listeners,
		ExecutionMode: execMode,
		RestartPolicy: config.RestartPolicyConfig{
			Policy:            "on-failure",
			MaxRestarts:       5,
			RestartDelay:      5 * time.Second,
			BackoffMultiplier: 2.0,
			MaxBackoffDelay:   5 * time.Minute,
		},
		HealthCheck: config.HealthCheckConfig{
			Enabled:  true,
			Interval: 10 * time.Second,
			Timeout:  5 * time.Second,
			Retries:  3,
		},
	}
}
