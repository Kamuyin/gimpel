export interface Module {
    id: string;
    name: string;
    description: string;
    version: string;
    protocol: string;
    digest: string;
    size_bytes: number;
    signed_by: string;
    signed_at: number;
    created_at: string;
}

export interface Satellite {
    id: string;
    hostname: string;
    ip_address: string;
    os: string;
    arch: string;
    status: string;
    registered_at: string;
    last_seen_at: string;
}

export interface Deployment {
    satellite_id: string;
    version: number;
    modules: ModuleAssignment[];
    updated_at: string;
}

export interface ModuleAssignment {
    module_id: string;
    module_version: string;
    execution_mode?: string;
    listeners?: Listener[];
    env?: Record<string, string>;
}

export interface Listener {
    id: string;
    protocol: string;
    port: number;
    high_interaction?: boolean;
}

export interface Pairing {
    id: string;
    display_token: string;
    created_at: string;
    expires_at: string;
    used: boolean;
    used_at?: string;
    assigned_agent?: string;
    agent_hostname?: string;
}

export default class APIService {
    private baseURL: string;

    constructor(baseURL: string = "/api/v1") {
        this.baseURL = baseURL;
    }

    // Generic fetch wrapper
    private async fetch<T>(endpoint: string, options?: RequestInit): Promise<T> {
        const url = `${this.baseURL}${endpoint}`;
        
        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    "Content-Type": "application/json",
                    ...options?.headers
                }
            });

            if (!response.ok) {
                throw new Error(`API Error: ${response.status} ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error(`API request failed: ${url}`, error);
            throw error;
        }
    }

    // Module Management 

    /**
     * List all modules
     */
    async listModules(): Promise<{ modules: Module[] }> {
        return this.fetch<{ modules: Module[] }>("/modules");
    }

    /**
     * Get module details
     */
    async getModule(id: string, version: string): Promise<Module> {
        return this.fetch<Module>(`/modules/${id}/${version}`);
    }

    /**
     * Delete a module
     */
    async deleteModule(id: string, version: string): Promise<{ status: string }> {
        return this.fetch<{ status: string }>(`/modules/${id}/${version}`, {
            method: "DELETE"
        });
    }

    /**
     * Upload a module
     */
    async uploadModule(formData: FormData): Promise<any> {
        const url = `${this.baseURL}/modules`;
        const response = await fetch(url, {
            method: "POST",
            body: formData
        });

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.status} ${response.statusText}`);
        }

        return await response.json();
    }

    // Satellite Management 

    /**
     * List all satellites
     */
    async listSatellites(): Promise<{ satellites: Satellite[] }> {
        return this.fetch<{ satellites: Satellite[] }>("/satellites");
    }

    /**
     * Get satellite details
     */
    async getSatellite(id: string): Promise<Satellite> {
        return this.fetch<Satellite>(`/satellites/${id}`);
    }

    // Deployment Management 

    /**
     * List all deployments
     */
    async listDeployments(): Promise<{ deployments: Deployment[] }> {
        return this.fetch<{ deployments: Deployment[] }>("/deployments");
    }

    /**
     * Get deployment for a satellite
     */
    async getDeployment(satelliteId: string): Promise<Deployment> {
        return this.fetch<Deployment>(`/satellites/${satelliteId}/deployments`);
    }

    /**
     * Create/update deployment for a satellite
     */
    async createDeployment(satelliteId: string, modules: ModuleAssignment[]): Promise<Deployment> {
        return this.fetch<Deployment>(`/satellites/${satelliteId}/deployments`, {
            method: "POST",
            body: JSON.stringify({ modules })
        });
    }

    /**
     * Delete deployment for a satellite
     */
    async deleteDeployment(satelliteId: string): Promise<{ status: string }> {
        return this.fetch<{ status: string }>(`/satellites/${satelliteId}/deployments`, {
            method: "DELETE"
        });
    }

    // Pairing Management 

    /**
     * Create a new pairing token
     */
    async createPairing(ttlSeconds?: number): Promise<any> {
        return this.fetch<any>("/pairings", {
            method: "POST",
            body: JSON.stringify({ ttl_seconds: ttlSeconds || 600 })
        });
    }

    /**
     * List all pairings
     */
    async listPairings(): Promise<{ pairings: Pairing[] }> {
        return this.fetch<{ pairings: Pairing[] }>("/pairings");
    }

    /**
     * List active pairings only
     */
    async listActivePairings(): Promise<{ pairings: Pairing[] }> {
        return this.fetch<{ pairings: Pairing[] }>("/pairings/active");
    }
}
