package providers

func SetupProcessor() *Processor {
	processor := NewProcessor()

	processor.RegisterPayoutProvider(NewCurrencyCloudProvider())
	processor.RegisterPayoutProvider(NewDLocalProvider())

	return processor
}
