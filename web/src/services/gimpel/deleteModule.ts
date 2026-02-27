// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Delete module DELETE /api/v1/modules/${param0}/${param1} */
export async function deleteModule(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteModuleParams,
  options?: { [key: string]: any },
) {
  const { id: param0, version: param1, ...queryParams } = params;
  return request<{ status?: string }>(`/api/v1/modules/${param0}/${param1}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}
