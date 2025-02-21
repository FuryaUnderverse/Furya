package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestSetHistoraclePricing() {
	app, ctx := s.app, s.ctx

	// update stamp params
	app.OracleKeeper.SetHistoricStampPeriod(ctx, 1)
	app.OracleKeeper.SetMedianStampPeriod(ctx, 3)

	// set historic prices and medians for denom with a similar
	// prefix to test unique prefix safety when iterating for
	// similar prefixes
	displayDenomVariation := displayDenom + "test"

	// add multiple historic prices to store
	exchangeRates := []string{"1.0", "1.2", "1.1", "1.4", "1.1", "1.15", "1.2", "1.3", "1.2"}
	for i, exchangeRate := range exchangeRates {
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, sdk.MustNewDecFromStr(exchangeRate))
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenomVariation, sdk.MustNewDecFromStr(exchangeRate))
		if ((i + 1) % int(app.OracleKeeper.MedianStampPeriod(ctx))) == 0 {
			err := app.OracleKeeper.CalcAndSetHistoricMedian(ctx, displayDenom)
			s.Require().NoError(err)
			err = app.OracleKeeper.CalcAndSetHistoricMedian(ctx, displayDenomVariation)
			s.Require().NoError(err)
		}

		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// check median and median standard deviation
	medians := app.OracleKeeper.HistoricMedians(ctx, displayDenom, 3)
	s.Require().Equal(len(medians), 3)
	s.Require().Equal(medians[0], sdk.MustNewDecFromStr("1.2"))
	s.Require().Equal(medians[1], sdk.MustNewDecFromStr("1.125"))
	s.Require().Equal(medians[2], sdk.MustNewDecFromStr("1.1"))

	medianDeviation, err := app.OracleKeeper.HistoricMedianDeviation(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(medianDeviation, sdk.MustNewDecFromStr("0.012499999999999998"))

	// check current price is within the median deviation
	result, err := app.OracleKeeper.WithinHistoricMedianDeviation(ctx, displayDenom)
	s.Require().Equal(result, true)
	s.Require().NoError(err)

	// check median stats of last 3 stamps
	medianOfMedians, err := app.OracleKeeper.MedianOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.125"))
	averageOfMedians, err := app.OracleKeeper.AverageOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.141666666666666666"))
	maxMedian, err := app.OracleKeeper.MaxOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.2"))
	minMedian, err := app.OracleKeeper.MinOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.1"))

	// check median stats of last 1 stamps
	medianOfMedians, err = app.OracleKeeper.MedianOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.2"))
	averageOfMedians, err = app.OracleKeeper.AverageOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.2"))
	maxMedian, err = app.OracleKeeper.MaxOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.2"))
	minMedian, err = app.OracleKeeper.MinOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.2"))

	// delete first median
	blockPeriod := (3 - 1) * app.OracleKeeper.MedianStampPeriod(ctx)
	lastStampBlock := uint64(ctx.BlockHeight()) - (uint64(ctx.BlockHeight())%app.OracleKeeper.MedianStampPeriod(ctx) + 1)
	firstStampBlock := lastStampBlock - blockPeriod
	app.OracleKeeper.DeleteHistoricMedian(ctx, displayDenom, firstStampBlock)

	medians = app.OracleKeeper.HistoricMedians(ctx, displayDenom, 3)
	s.Require().Equal(len(medians), 2)
	s.Require().Equal(medians[0], sdk.MustNewDecFromStr("1.2"))
	s.Require().Equal(medians[1], sdk.MustNewDecFromStr("1.125"))
}
