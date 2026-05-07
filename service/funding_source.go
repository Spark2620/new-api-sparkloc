package service

import (
	"time"

	"github.com/QuantumNous/new-api/model"
)

type FundingSource interface {
	Source() string
	PreConsume(amount int) error
	Settle(delta int) error
	Refund() error
}

type WalletFunding struct {
	userId         int
	requestId      string
	dailyConsumed  int
	earnedConsumed int
	legacyConsumed int
}

func (w *WalletFunding) Source() string { return BillingSourceWallet }

func (w *WalletFunding) PreConsume(amount int) error {
	return w.consume(amount)
}

func (w *WalletFunding) consume(amount int) error {
	if amount <= 0 {
		return nil
	}
	community, err := model.ConsumeCommunityCredits(w.requestId, w.userId, amount)
	if err != nil {
		return err
	}
	legacyAmount := amount
	if community != nil {
		legacyAmount = community.Shortage
	}
	if legacyAmount > 0 {
		if err := model.DecreaseUserQuota(w.userId, legacyAmount, false); err != nil {
			if community != nil && community.TotalAmount > 0 {
				_, _ = model.RefundCommunityCredits(w.requestId, w.userId, community.TotalAmount)
			}
			return err
		}
		w.legacyConsumed += legacyAmount
	}
	if community != nil {
		w.dailyConsumed += community.DailyAmount
		w.earnedConsumed += community.EarnedAmount
	}
	return nil
}

func (w *WalletFunding) Settle(delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return w.consume(delta)
	}
	return w.refund(-delta)
}

func (w *WalletFunding) Refund() error {
	return w.refund(w.TotalConsumed())
}

func (w *WalletFunding) refund(amount int) error {
	if amount <= 0 {
		return nil
	}
	remaining := amount
	if w.legacyConsumed > 0 {
		refundLegacy := w.legacyConsumed
		if refundLegacy > remaining {
			refundLegacy = remaining
		}
		if refundLegacy > 0 {
			if err := model.IncreaseUserQuota(w.userId, refundLegacy, false); err != nil {
				return err
			}
			w.legacyConsumed -= refundLegacy
			remaining -= refundLegacy
		}
	}
	if remaining <= 0 {
		return nil
	}
	community, err := model.RefundCommunityCredits(w.requestId, w.userId, remaining)
	if err != nil {
		return err
	}
	if community != nil {
		w.dailyConsumed -= community.DailyAmount
		if w.dailyConsumed < 0 {
			w.dailyConsumed = 0
		}
		w.earnedConsumed -= community.EarnedAmount
		if w.earnedConsumed < 0 {
			w.earnedConsumed = 0
		}
	}
	return nil
}

func (w *WalletFunding) TotalConsumed() int {
	return w.dailyConsumed + w.earnedConsumed + w.legacyConsumed
}

func (w *WalletFunding) DailyConsumed() int {
	return w.dailyConsumed
}

func (w *WalletFunding) EarnedConsumed() int {
	return w.earnedConsumed
}

func (w *WalletFunding) LegacyConsumed() int {
	return w.legacyConsumed
}

type SubscriptionFunding struct {
	requestId      string
	userId         int
	modelName      string
	amount         int64
	subscriptionId int
	preConsumed    int64

	AmountTotal     int64
	AmountUsedAfter int64
	PlanId          int
	PlanTitle       string
}

func (s *SubscriptionFunding) Source() string { return BillingSourceSubscription }

func (s *SubscriptionFunding) PreConsume(_ int) error {
	res, err := model.PreConsumeUserSubscription(s.requestId, s.userId, s.modelName, 0, s.amount)
	if err != nil {
		return err
	}
	s.subscriptionId = res.UserSubscriptionId
	s.preConsumed = res.PreConsumed
	s.AmountTotal = res.AmountTotal
	s.AmountUsedAfter = res.AmountUsedAfter
	if planInfo, err := model.GetSubscriptionPlanInfoByUserSubscriptionId(res.UserSubscriptionId); err == nil && planInfo != nil {
		s.PlanId = planInfo.PlanId
		s.PlanTitle = planInfo.PlanTitle
	}
	return nil
}

func (s *SubscriptionFunding) Settle(delta int) error {
	if delta == 0 {
		return nil
	}
	return model.PostConsumeUserSubscriptionDelta(s.subscriptionId, int64(delta))
}

func (s *SubscriptionFunding) Refund() error {
	if s.preConsumed <= 0 {
		return nil
	}
	return refundWithRetry(func() error {
		return model.RefundSubscriptionPreConsume(s.requestId)
	})
}

func refundWithRetry(fn func() error) error {
	if fn == nil {
		return nil
	}
	const maxAttempts = 3
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if i < maxAttempts-1 {
			time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
		}
	}
	return lastErr
}
