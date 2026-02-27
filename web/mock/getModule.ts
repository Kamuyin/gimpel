// @ts-ignore
import { Request, Response } from 'express';

export default {
  'GET /api/v1/modules/:id/:version': (req: Request, res: Response) => {
    res.status(200).send({
      id: 'fe4253f0-CFC8-470b-5e4E-8bEAc7ab2c9D',
      name: '黎平',
      description: '年社月按政它之影四打效音。',
      version: '青分步界商着传等传经养布什。',
      protocol: '对青精代以类科青业铁入子较。',
      digest: '类场龙头已需向节响金建象精动社连况。',
      size_bytes: 95,
      signed_by: '已取研自较工今族治特知型非此以。',
      signed_at: 68,
      created_at: '2022-11-24 13:40:50',
    });
  },
};
