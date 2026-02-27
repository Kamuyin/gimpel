// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Get satellite GET /api/v1/satellites/${param0} */
export async function getSatellite(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getSatelliteParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.SatelliteInfo>(`/api/v1/satellites/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}
