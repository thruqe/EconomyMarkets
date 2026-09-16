package taxation

// Taxation models the citizen tax framework in an American-style system.
type Taxation struct {
	EffectiveIncomeTaxRate float64 // Average effective federal income tax rate (~18.5%)
	CapitalGainsTaxRate    float64 // Long-term capital gains tax rate (~15.0%)
	SalesTaxRate           float64 // Average state & local consumption tax rate (~6.5%)
	TotalIncomeTaxesPaid   float64 // Cumulative USD collected from personal income
	TotalCapGainsTaxesPaid float64 // Cumulative USD collected from capital gains
	TotalSalesTaxesPaid    float64 // Cumulative USD collected from sales
}

// NewDefaultTaxation initializes standard federal and state citizen tax rates.
func NewDefaultTaxation() *Taxation {
	return &Taxation{
		EffectiveIncomeTaxRate: 0.185,
		CapitalGainsTaxRate:    0.150,
		SalesTaxRate:           0.065,
	}
}

// ComputeIncomeTax calculates the tax owed on citizen gross earnings.
func (t *Taxation) ComputeIncomeTax(grossIncome float64) float64 {
	return grossIncome * t.EffectiveIncomeTaxRate
}

// ComputeSalesTax calculates consumption tax on retail spending.
func (t *Taxation) ComputeSalesTax(consumerSpending float64) float64 {
	return consumerSpending * t.SalesTaxRate
}

// ComputeCapitalGainsTax calculates tax on realized market profits.
func (t *Taxation) ComputeCapitalGainsTax(realizedGains float64) float64 {
	if realizedGains <= 0 {
		return 0.0
	}
	return realizedGains * t.CapitalGainsTaxRate
}

// RecordTaxes adds collected receipts to cumulative tallies.
func (t *Taxation) RecordTaxes(incomeTax, capGainsTax, salesTax float64) {
	t.TotalIncomeTaxesPaid += incomeTax
	t.TotalCapGainsTaxesPaid += capGainsTax
	t.TotalSalesTaxesPaid += salesTax
}

// TotalCollected returns total cumulative citizen taxes paid to government.
func (t *Taxation) TotalCollected() float64 {
	return t.TotalIncomeTaxesPaid + t.TotalCapGainsTaxesPaid + t.TotalSalesTaxesPaid
}
