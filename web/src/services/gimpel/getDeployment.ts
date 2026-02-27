// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Get deployment GET /api/v1/satellites/${param0}/deployments */
export async function getDeployment(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDeploymentParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.DeploymentResponse>(
    `/api/v1/satellites/${param0}/deployments`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    },
  );
}
