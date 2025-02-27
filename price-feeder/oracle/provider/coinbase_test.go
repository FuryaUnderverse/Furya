package provider

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/v2/oracle/types"
)

func TestCoinbaseProvider_GetTickerPrices(t *testing.T) {
	p, err := NewCoinbaseProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "BTC", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := "34.69000000"
		volume := "2396974.02000000"

		tickerMap := map[string]CoinbaseTicker{}
		tickerMap["ATOM-USDT"] = CoinbaseTicker{
			Price:  lastPrice,
			Volume: volume,
		}

		p.tickers = tickerMap

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr(lastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		lastPriceAtom := "34.69000000"
		lastPriceUmee := "41.35000000"
		volume := "2396974.02000000"

		tickerMap := map[string]CoinbaseTicker{}
		tickerMap["ATOM-USDT"] = CoinbaseTicker{
			Price:  lastPriceAtom,
			Volume: volume,
		}

		tickerMap["UMEE-USDT"] = CoinbaseTicker{
			Price:  lastPriceUmee,
			Volume: volume,
		}

		p.tickers = tickerMap
		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "UMEE", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr(lastPriceAtom), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr(lastPriceUmee), prices["UMEEUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["UMEEUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "coinbase failed to get ticker price for FOO-BAR")
		require.Nil(t, prices)
	})
}

func TestCoinbasePairToCurrencyPair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	currencyPairSymbol := coinbasePairToCurrencyPair("ATOM-USDT")
	require.Equal(t, cp.String(), currencyPairSymbol)
}

func TestCurrencyPairToCoinbasePair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	coinbaseSymbol := currencyPairToCoinbasePair(cp)
	require.Equal(t, coinbaseSymbol, "ATOM-USDT")
}

func TestCoinbaseProvider_getSubscriptionMsgs(t *testing.T) {
	provider := &CoinbaseProvider{
		subscribedPairs: map[string]types.CurrencyPair{},
	}
	cps := []types.CurrencyPair{
		{Base: "ATOM", Quote: "USDT"},
	}
	subMsgs := provider.getSubscriptionMsgs(cps...)

	msg, _ := json.Marshal(subMsgs[0])
	require.Equal(t, "{\"type\":\"subscribe\",\"product_ids\":[\"ATOM-USDT\"],\"channels\":[\"matches\",\"ticker\"]}", string(msg))
}
