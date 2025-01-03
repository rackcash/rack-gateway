
import { writeFileSync } from "fs";
import { Telegraf } from "telegraf";

const BOT_TOKEN = "";
const RACK_URL = "http://127.0.0.1:8888/v1";
const RACK_API_KEY = "ae9f3181cac51f65692e0e8d5f71c5a2042700c421caa595926aab777a391d55";

const bot = new Telegraf(BOT_TOKEN);

let info = await bot.telegram.getMe();
console.log("Started: https://t.me/" + info.username);

bot.start(async (ctx) => {
    const url = RACK_URL + "/invoice/create";
    const data = {
        lifetime: 10, // 10 minutes
        amount: 0.001437, // 1 ETH
        api_key: RACK_API_KEY,
        cryptocurrency: "eth",
    };

    const response = await fetch(url, { method: "POST", body: JSON.stringify(data) });
    const result = await response.json();

    const qrCode = await fetch(result.invoice.wallet.qr_code).then(res => res.arrayBuffer());
    writeFileSync("qr_code.png", Buffer.from(qrCode));

    await ctx.replyWithPhoto({ source: "qr_code.png" }, {
        caption: `Invoice ID: ${result.invoice.id}\nSend ${result.invoice.wallet.amount_to_pay} ${result.invoice.wallet.cryptocurrency} to ${result.invoice.wallet.address} `, reply_markup: {
            inline_keyboard: [
                [
                    { text: "Check invoice", callback_data: "check_tx:" + result.invoice.id },
                ]
            ]
        }
    });
});


bot.action(/check_tx:(.+)/, async (ctx) => {
    ctx.answerCbQuery();

    const invoice_id = ctx.match[1];

    const url = RACK_URL + "/invoice/info";
    const data = {
        invoice_id: invoice_id,
        api_key: RACK_API_KEY,
    };

    const response = await fetch(url, { method: "POST", body: JSON.stringify(data) });
    const result = await response.json();

    if (result.is_paid) {
        return await ctx.reply("paid");
    }

    return await ctx.reply("not paid");
});


bot.catch(err => {
    console.log("Bot Catch: " + err);
});

bot.launch({ dropPendingUpdates: true });
