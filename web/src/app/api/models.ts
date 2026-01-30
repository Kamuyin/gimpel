export interface ModuleInfo {
    id: string;
    name?: string;
    description?: string;
    version: string;
    protocol?: string;
    digest?: string;
    size_bytes?: number;
    signed_by?: string;
    signed_at?: number;
    created_at?: string;
}

export interface ListModulesResponse {
    modules: ModuleInfo[];
}

export interface Listener {
    id: string;
    protocol: string;
    port: number;
    high_interaction?: boolean;
}

export interface ModuleAssignment {
    module_id: string;
    module_version: string;
    execution_mode?: string;
    listeners?: Listener[];
    env?: Record<string, string>;
}

export interface CreateDeploymentRequest {
    modules: ModuleAssignment[];
}

export interface DeploymentResponse {
    satellite_id: string;
    version: number;
    modules: ModuleAssignment[];
    updated_at?: string;
}

export interface SatelliteInfo {
    id: string;
    hostname?: string;
    ip_address?: string;
    os?: string;
    arch?: string;
    status?: string;
    registered_at?: string;
    last_seen_at?: string;
}

export interface ListSatellitesResponse {
    satellites: SatelliteInfo[];
}

export interface CreatePairingRequest {
    ttl_seconds?: number;
}

export interface PairingResponse {
    id: string;
    token: string;
    display_token: string;
    expires_at?: string;
    expires_in_seconds?: number;
}

export interface PairingInfo {
    id: string;
    display_token: string;
    created_at?: string;
    expires_at?: string;
    used?: boolean;
    used_at?: string;
    assigned_agent?: string;
    agent_hostname?: string;
}

export interface ListPairingsResponse {
    pairings: PairingInfo[];
}
