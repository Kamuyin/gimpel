// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/v1/pairings': (req: Request, res: Response) => {
    res.status(200).send({
      id: '938b9BFe-8ACb-6B6F-Ee7C-C30cf19B76Bc',
      token: '格在片六已石二改克导影料进精月新。',
      display_token: '难引火指传决变准基改精称单。',
      expires_at: '1981-05-25 08:46:45',
      expires_in_seconds: 81,
    });
  },
};
