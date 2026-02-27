// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** List deployments GET /api/v1/deployments */
export async function listDeployments(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listDeploymentsParams,
  options?: { [key: string]: any },
) {
  return request<{ deployments?: API.DeploymentResponse[] }>(
    '/api/v1/deployments',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    },
  );
}
