// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** List modules GET /api/v1/modules */
export async function listModules(options?: { [key: string]: any }) {
  return request<API.ListModulesResponse>('/api/v1/modules', {
    method: 'GET',
    ...(options || {}),
  });
}
