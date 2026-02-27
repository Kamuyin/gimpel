declare namespace API {
  type createDeploymentParams = {
    id: string;
  };

  type CreateDeploymentRequest = {
    modules?: ModuleAssignment[];
  };

  type CreatePairingRequest = {
    ttl_seconds?: number;
  };

  type deleteDeploymentParams = {
    id: string;
  };

  type deleteModuleParams = {
    id: string;
    version: string;
  };

  type DeploymentResponse = {
    satellite_id?: string;
    version?: number;
    modules?: ModuleAssignment[];
    updated_at?: string;
  };

  type downloadModuleParams = {
    id: string;
    version: string;
  };

  type getDeploymentParams = {
    id: string;
  };

  type getModuleParams = {
    id: string;
    version: string;
  };

  type getSatelliteParams = {
    id: string;
  };

  type listDeploymentsParams = {
    status?: string;
  };

  type Listener = {
    id?: string;
    protocol?: string;
    port?: number;
    high_interaction?: boolean;
  };

  type ListModulesResponse = {
    modules?: ModuleInfo[];
  };

  type ListPairingsResponse = {
    pairings?: PairingInfo[];
  };

  type ListSatellitesResponse = {
    satellites?: SatelliteInfo[];
  };

  type ModuleAssignment = {
    module_id?: string;
    module_version?: string;
    execution_mode?: string;
    listeners?: Listener[];
    env?: Record<string, any>;
  };

  type ModuleInfo = {
    id?: string;
    name?: string;
    description?: string;
    version?: string;
    protocol?: string;
    digest?: string;
    size_bytes?: number;
    signed_by?: string;
    signed_at?: number;
    created_at?: string;
  };

  type PairingInfo = {
    id?: string;
    display_token?: string;
    created_at?: string;
    expires_at?: string;
    used?: boolean;
    used_at?: string;
    assigned_agent?: string;
    agent_hostname?: string;
  };

  type PairingResponse = {
    id?: string;
    token?: string;
    display_token?: string;
    expires_at?: string;
    expires_in_seconds?: number;
  };

  type SatelliteInfo = {
    id?: string;
    hostname?: string;
    ip_address?: string;
    os?: string;
    arch?: string;
    status?: string;
    registered_at?: string;
    last_seen_at?: string;
  };

  type UploadModuleResponse = {
    id?: string;
    version?: string;
    digest?: string;
    signature?: string;
    signed_by?: string;
    signed_at?: number;
    size?: number;
    created_at?: string;
  };
}
