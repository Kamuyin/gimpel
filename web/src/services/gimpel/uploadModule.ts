// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Upload module POST /api/v1/modules */
export async function uploadModule(
  body: {
    id: string;
    name?: string;
    description?: string;
    version: string;
    protocol?: string;
    /** Hex-encoded signature */
    signature: string;
    signed_at?: number;
  },
  image?: File,
  options?: { [key: string]: any },
) {
  const formData = new FormData();

  if (image) {
    formData.append('image', image);
  }

  Object.keys(body).forEach((ele) => {
    const item = (body as any)[ele];

    if (item !== undefined && item !== null) {
      if (typeof item === 'object' && !(item instanceof File)) {
        if (item instanceof Array) {
          item.forEach((f) => formData.append(ele, f || ''));
        } else {
          formData.append(
            ele,
            new Blob([JSON.stringify(item)], { type: 'application/json' }),
          );
        }
      } else {
        formData.append(ele, item);
      }
    }
  });

  return request<API.UploadModuleResponse>('/api/v1/modules', {
    method: 'POST',
    data: formData,
    requestType: 'form',
    ...(options || {}),
  });
}
