package providers

func SetupProcessor() *Processor {
	processor := NewProcessor()

	ccProvider := NewCurrencyCloudProvider()
	dlocalProvider := NewDLocalProvider()

	processor.RegisterPayoutProvider(ccProvider)
	processor.RegisterPayoutProvider(dlocalProvider)

	processor.RegisterNameEnquiryProvider(ccProvider)
	processor.RegisterNameEnquiryProvider(dlocalProvider)

	processor.RegisterExchangeRateProvider(ccProvider)
	processor.RegisterExchangeRateProvider(dlocalProvider)

	return processor
}
