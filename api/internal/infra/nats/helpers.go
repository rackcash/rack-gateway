package nats

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"log"
)

// checks if there is an error in the response. if there is, it returns true and the error message
func HelpersIsError(data []byte) (bool, string) {
	if len(data) < 6 {
		return false, ""
	}

	if string(data[0:6]) == "error:" {
		return true, string(data[6:])

	}
	return false, ""
}

// server withdrawal returns tx hash.
// example:
// ok: 0x1234567890123456789012345678901234567890123456789012345678901234
func HelpersInvoiceGetTxHash(data []byte) (string, error) {
	if len(data) < 10 {
		return "", nil
	}

	if string(data[0:3]) != "ok:" {
		return "", fmt.Errorf("data[0:3] is not 'ok:': " + string(data))
	}

	return string(data[4:]), nil
}

// get balance wrapper
//
//	crypto - cryptocurrency (eth, sol, ton)
//	address - address from which the balance is received
func (n *NatsInfra) ReqGetBalance(crypto string, address string) (*natsdomain.ResGetBalance, error) {

	data, err := json.Marshal(natsdomain.ReqGetBalance{Cryptocurrency: crypto, Address: address})
	if err != nil {
		return nil, err
	}

	resp, err := n.Ns.ReqAndRecv(natsdomain.SubjGetBalance, data)
	if err != nil {
		return nil, err
	}

	iserr, errmsg := HelpersIsError(resp)
	if iserr {
		return nil, fmt.Errorf(errmsg)
	}

	balance, err := utils.Unmarshal[natsdomain.ResGetBalance](resp)
	if err != nil {
		return nil, err
	}

	return balance, nil

}

func (n *NatsInfra) ReqIsPaid(invoice *domain.Invoices, address string) (*natsdomain.ResIsPaid, error) {
	reqData, err := json.Marshal(natsdomain.ReqIsPaid{Address: address, Cryptocurrency: invoice.Cryptocurrency, TargetAmount: invoice.Amount})
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	log.Println("check 2")

	resp, err := n.ReqAndRecv(natsdomain.SubjIsPayed, reqData)
	if err != nil {
		return nil, fmt.Errorf("reqAndRecv error: %w", err)
	}

	isError, errmsg := HelpersIsError(resp)
	if isError {
		return nil, fmt.Errorf("error in is paid response: %s", errmsg)
	}

	msg, err := utils.Unmarshal[natsdomain.ResIsPaid](resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return msg, nil

}

func (n *NatsInfra) ReqWithdrawal(invoice *domain.Invoices, wallet *domain.Wallets, balance *domain.Balances, status string) error {
	data, err := json.Marshal(natsdomain.ReqWithdrawal{InvoiceId: invoice.InvoiceID, MerchantId: invoice.MerchantID, Address: balance.Address, Private: wallet.Private, Crypto: wallet.Crypto, Amount: wallet.Balance, Status: status, TxTempWallet: wallet.Address})
	if err != nil {
		return err
	}

	err = n.JsPublishMsgId(natsdomain.SubjJsWithdraw.String(), data, natsdomain.NewMsgId(invoice.InvoiceID, natsdomain.MsgActionWithdrawal))
	if err != nil {
		return err
	}
	return nil
}
