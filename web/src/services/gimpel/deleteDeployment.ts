// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Delete deployment DELETE /api/v1/satellites/${param0}/deployments */
export async function deleteDeployment(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteDeploymentParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<{ status?: string }>(
    `/api/v1/satellites/${param0}/deployments`,
    {
      method: 'DELETE',
      params: { ...queryParams },
      ...(options || {}),
    },
  );
}
