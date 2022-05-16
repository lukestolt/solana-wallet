package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

// referenced doc https://blog.logrocket.com/how-to-create-solana-wallet-go/

type Wallet struct {
	account types.Account
	c       *client.Client
}

func CreateNewWallet(RPCEndPoint string) Wallet {
	return Wallet{
		types.NewAccount(),
		client.NewClient(RPCEndPoint),
	}
}

func ImportOldWallet(privateKey []byte, RPCEndPoint string) (Wallet, error) {
	wallet, err := types.AccountFromBytes(privateKey)
	if err != nil {
		return Wallet{}, err
	}

	return Wallet{
		wallet,
		client.NewClient(RPCEndPoint),
	}, nil
}

func (w Wallet) RequestAirdrop(lamports uint64) (string, error) {
	tx, err := w.c.RequestAirdrop(context.TODO(), w.account.PublicKey.ToBase58(), lamports)
	if err != nil {
		return "", err
	}
	return tx, nil
}

func (w Wallet) GetBalance() (uint64, error) {
	balance, err := w.c.GetBalance(context.TODO(), w.account.PublicKey.ToBase58())
	if err != nil {
		return 0, nil
	}
	return balance, nil
}

func (w Wallet) Transfer(reciever string, lamports uint64) (string, error) {
	resp, err := w.c.GetLatestBlockhash(context.TODO())
	if err != nil {
		return "get hash fail", err
	}

	message := types.NewMessage(types.NewMessageParam{
		w.account.PublicKey,
		[]types.Instruction{
			sysprog.Transfer(sysprog.TransferParam{w.account.PublicKey, common.PublicKeyFromString(reciever), lamports}),
		},
		resp.Blockhash,
	},
	)
	tx, err := types.NewTransaction(types.NewTransactionParam{message, []types.Account{w.account, w.account}})
	if err != nil {
		return "tx fail", err
	}

	failCount := 0

	for failCount < 10 {
		_, err := w.c.SendTransaction(context.TODO(), tx)
		if err != nil {
			failCount++
			fmt.Println("Fail")
		} else {
			return fmt.Sprintf("Took %d seconds to send tx", failCount+1*2), nil
		}
		time.Sleep(2 * time.Second)

	}
	return "send tx fail", nil

}

func GetWalletBalance(publicKey string) (uint64, error) {
	c := client.NewClient(rpc.DevnetRPCEndpoint)
	return c.GetBalance(context.TODO(), publicKey)
}

func main() {
	// import a wallet or create a new wallet

	// create a new wallet with 2 SOL
	wallet := CreateNewWallet(rpc.DevnetRPCEndpoint)
	fmt.Println("New Wallet Public key ", wallet.account.PublicKey.String())
	fmt.Println("New Wallet Private key ", wallet.account.PrivateKey)
	fmt.Println(wallet.RequestAirdrop(2e9))
	fmt.Println("Recieving Airdrop.....")
	time.Sleep(10 * time.Second)
	// address to send SOL to
	fmt.Print("\nAddress to send SOL to: ")
	var recieverAddr string
	fmt.Scan(&recieverAddr)
	// amount of SOL to send
	fmt.Print("Amount of lamports to send: ")
	var lamports uint64
	fmt.Scan(&lamports)
	// check that the wallet has enough sol to send
	balance, _ := wallet.GetBalance()
	if lamports > balance {
		fmt.Println("ERROR: wallet balance: " + strconv.FormatUint(balance, 10) +
			" is < the amount requested to be sent: " + strconv.FormatUint(lamports, 10))
	}
	// send the tx
	wallet.Transfer(recieverAddr, lamports)
	// verify SOL was recieved
	curBalance, _ := wallet.GetBalance()
	otherWalletBalance, _ := GetWalletBalance(recieverAddr)
	fmt.Println("Your wallets balance: " + strconv.FormatUint(curBalance, 10))
	fmt.Println("Other wallets balance: " + strconv.FormatUint(otherWalletBalance, 10))

}
