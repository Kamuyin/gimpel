// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Create pairing token POST /api/v1/pairings */
export async function createPairing(
  body: API.CreatePairingRequest,
  options?: { [key: string]: any },
) {
  return request<API.PairingResponse>('/api/v1/pairings', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
