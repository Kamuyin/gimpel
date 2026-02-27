// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** List active pairings GET /api/v1/pairings/active */
export async function listActivePairings(options?: { [key: string]: any }) {
  return request<API.ListPairingsResponse>('/api/v1/pairings/active', {
    method: 'GET',
    ...(options || {}),
  });
}
