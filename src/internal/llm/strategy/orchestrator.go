package strategy

type StrategyOrchestrator struct {
    strategies []OrganizationStrategy
    selector   StrategySelector
}

type OrganizationStrategy interface {
    Analyze(ctx context.Context, file FileData) (DestinationDecision, float64)
    IsApplicable(file FileData) bool
}

type StrategySelector struct {
    weights map[string]float64
}

func (o *StrategyOrchestrator) ProcessFile(ctx context.Context, file FileData) DestinationDecision {
    applicable := o.findApplicableStrategies(file)
    decisions := make([]WeightedDecision, 0)
    
    for _, strategy := range applicable {
        decision, confidence := strategy.Analyze(ctx, file)
        weight := o.selector.GetWeight(strategy)
        decisions = append(decisions, WeightedDecision{
            Decision:   decision,
            Weight:    weight * confidence,
        })
    }
    
    return o.combineDecisions(decisions)
}