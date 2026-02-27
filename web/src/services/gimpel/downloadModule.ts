// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Download module image GET /api/v1/modules/${param0}/${param1}/download */
export async function downloadModule(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.downloadModuleParams,
  options?: { [key: string]: any },
) {
  const { id: param0, version: param1, ...queryParams } = params;
  return request<string>(`/api/v1/modules/${param0}/${param1}/download`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}
