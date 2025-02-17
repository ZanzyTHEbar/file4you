package learning

type FeedbackLoop struct {
    store      FeedbackStore
    classifier *DecisionClassifier
}

type UserFeedback struct {
    Decision     DestinationDecision
    Accepted     bool
    Alternative  *DestinationDecision
    Reason       string
}

func (f *FeedbackLoop) Learn(feedback UserFeedback) error {
    // Update decision classifier based on feedback
    features := extractFeatures(feedback.Decision)
    if feedback.Accepted {
        f.classifier.TrainPositive(features)
    } else {
        f.classifier.TrainNegative(features)
        if feedback.Alternative != nil {
            f.classifier.TrainPositive(extractFeatures(*feedback.Alternative))
        }
    }
    return f.store.SaveFeedback(feedback)
}

func (f *FeedbackLoop) AdjustDecision(decision DestinationDecision) DestinationDecision {
    features := extractFeatures(decision)
    confidence := f.classifier.Predict(features)
    
    if confidence < 0.7 {
        return f.findAlternative(decision)
    }
    return decision
}