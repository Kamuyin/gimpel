// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/v1/modules': (req: Request, res: Response) => {
    res.status(200).send({
      id: '7F0b144D-dAAf-9B78-FDc1-ce1bfA4dc58e',
      version: '一可今持名别二算被次构料。',
      digest: '式流积律品书解门了律委则观共众。',
      signature: '何传军响选表接节属除准但照队矿米引。',
      signed_by: '办过值府始制清明大节两会争式严报。',
      signed_at: 65,
      size: 88,
      created_at: '2003-03-13 10:06:06',
    });
  },
};
