// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** List satellites GET /api/v1/satellites */
export async function listSatellites(options?: { [key: string]: any }) {
  return request<API.ListSatellitesResponse>('/api/v1/satellites', {
    method: 'GET',
    ...(options || {}),
  });
}
