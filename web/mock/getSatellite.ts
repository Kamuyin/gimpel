// @ts-ignore
import { Request, Response } from 'express';

export default {
  'GET /api/v1/satellites/:id': (req: Request, res: Response) => {
    res.status(200).send({
      id: '36dC83c2-afcA-F237-23f2-F1176e2Be6F5',
      hostname: '打叫着人属目装八例己物联院发。',
      ip_address: '口线一型想次红对阶铁小效。',
      os: '书书发米单进属统始调见同格开。',
      arch: '同好斗往八等的则分劳样决研他者龙毛。',
      status: 'success',
      registered_at: '1983-07-30 07:02:14',
      last_seen_at: '1981-12-12 09:32:03',
    });
  },
};
