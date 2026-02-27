// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Get module GET /api/v1/modules/${param0}/${param1} */
export async function getModule(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getModuleParams,
  options?: { [key: string]: any },
) {
  const { id: param0, version: param1, ...queryParams } = params;
  return request<API.ModuleInfo>(`/api/v1/modules/${param0}/${param1}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}
