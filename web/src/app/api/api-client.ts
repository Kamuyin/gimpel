import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { environment } from '../../environments/environment';
import type {
    CreateDeploymentRequest,
    CreatePairingRequest,
    DeploymentResponse,
    ListModulesResponse,
    ListPairingsResponse,
    ListSatellitesResponse,
    ModuleInfo,
    PairingResponse,
} from './models';

@Injectable({ providedIn: 'root' })
export class ApiClient {
    private readonly http = inject(HttpClient);
    private readonly baseUrl = environment.apiBaseUrl.replace(/\/$/, '');

    listModules() {
        return this.http.get<ListModulesResponse>(`${this.baseUrl}/api/v1/modules`);
    }

    getModule(id: string, version: string) {
        return this.http.get<ModuleInfo>(`${this.baseUrl}/api/v1/modules/${id}/${version}`);
    }

    deleteModule(id: string, version: string) {
        return this.http.delete<{ status: string }>(`${this.baseUrl}/api/v1/modules/${id}/${version}`);
    }

    downloadModule(id: string, version: string) {
        return this.http.get(`${this.baseUrl}/api/v1/modules/${id}/${version}/download`, {
            responseType: 'blob'
        });
    }

    listSatellites() {
        return this.http.get<ListSatellitesResponse>(`${this.baseUrl}/api/v1/satellites`);
    }

    getSatellite(id: string) {
        return this.http.get(`${this.baseUrl}/api/v1/satellites/${id}`);
    }

    createDeployment(satelliteId: string, payload: CreateDeploymentRequest) {
        return this.http.post<DeploymentResponse>(`${this.baseUrl}/api/v1/satellites/${satelliteId}/deployments`, payload);
    }

    getDeployment(satelliteId: string) {
        return this.http.get<DeploymentResponse>(`${this.baseUrl}/api/v1/satellites/${satelliteId}/deployments`);
    }

    deleteDeployment(satelliteId: string) {
        return this.http.delete<{ status: string }>(`${this.baseUrl}/api/v1/satellites/${satelliteId}/deployments`);
    }

    listDeployments(status?: string) {
        let params = new HttpParams();
        if (status) {
            params = params.set('status', status);
        }
        return this.http.get<{ deployments: DeploymentResponse[] }>(`${this.baseUrl}/api/v1/deployments`, { params });
    }

    createPairing(payload?: CreatePairingRequest) {
        return this.http.post<PairingResponse>(`${this.baseUrl}/api/v1/pairings`, payload ?? {});
    }

    listPairings() {
        return this.http.get<ListPairingsResponse>(`${this.baseUrl}/api/v1/pairings`);
    }

    listActivePairings() {
        return this.http.get<ListPairingsResponse>(`${this.baseUrl}/api/v1/pairings/active`);
    }
}
