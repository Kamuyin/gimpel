// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Create deployment POST /api/v1/satellites/${param0}/deployments */
export async function createDeployment(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.createDeploymentParams,
  body: API.CreateDeploymentRequest,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.DeploymentResponse>(
    `/api/v1/satellites/${param0}/deployments`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    },
  );
}
