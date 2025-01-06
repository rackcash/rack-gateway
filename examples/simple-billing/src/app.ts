import express, { type Request, type Response } from 'express';
import path from 'path';


const app = express();
const PORT = 3000;

app.set('view engine', 'ejs');


const API_URL = "http://127.0.0.1:8888/v1";
const API_KEY = "f176ab8c801bf7576e6944fff0f24b3477087ad8f664e2a0bd811d0230d9a82a";
const LIFETIME = 10;

app.use(express.static(path.join(__dirname, '../public')));
app.use(express.json());
app.use(express.urlencoded({ extended: true }));


app.get("/", async (req: Request, res: Response) => {
  res.render('index');
});


app.post("/create", async (req: Request, res: Response) => {
  console.log(req.body);

  const body = {
    amount: Number(req.body.amount),
    lifetime: LIFETIME,
    api_key: API_KEY,
    cryptocurrency: req.body.cryptocurrency,
  };

  try {
    const response = await fetch(API_URL + '/invoice/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });


    if (!response.ok) {
      throw response.statusText;
    }

    const json = await response.json() as any;

    const qrCode = await fetch(json.invoice.wallet.qr_code).then(res => res.arrayBuffer());

    const encode = (str: any): string => Buffer.from(str, 'binary').toString('base64');
    const qrCodeData = `data:image/png;base64,${encode(qrCode)}`;

    res.render('check', {
      invoice_id: json.invoice.id,
      amount: json.invoice.wallet.amount_to_pay,
      cryptocurrency: json.invoice.wallet.cryptocurrency,
      address: json.invoice.wallet.address,
      qr_code: qrCodeData
    });

  } catch (error) {
    console.log(error);
    return;
  }
});



app.post("/check", async (req: Request, res: Response) => {
  const body = {
    invoice_id: req.body.invoice_id,
    api_key: API_KEY,
  };


  try {
    const response = await fetch(API_URL + '/invoice/info', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });


    if (!response.ok) {
      throw response.statusText;
    }

    const json = await response.json() as any;

    res.json(json);
  } catch (error) {
    console.log(error);
    return;
  }
});


app.listen(PORT, () => {
  console.log(`Started http://localhost:${PORT}`);
});
