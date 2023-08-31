package cpubbr

import (
	aegislimiter "github.com/go-kratos/aegis/ratelimit"
	"github.com/go-kratos/aegis/ratelimit/bbr"
	"github.com/polarismesh/polaris-go/pkg/model"
	"github.com/polarismesh/polaris-go/pkg/model/pb"
	"github.com/polarismesh/polaris-go/pkg/plugin/ratelimiter"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"
)

type CpuBbrQuota struct {
	aegislimiter.Limiter
}

// GetQuota 获取限额
func (b *CpuBbrQuota) GetQuota(_ int64, _ uint32) *model.QuotaResponse {
	// 默认返回 ok
	resp := &model.QuotaResponse{Code: model.QuotaResultOk}

	// 如果触发限流，err 值将等于 aegislimiter.ErrLimitExceed
	done, err := b.Limiter.Allow()
	if err != nil {
		resp.Code = model.QuotaResultLimited
		return resp
	}

	// 返回函数执行
	done(aegislimiter.DoneInfo{})

	return resp
}

// Release 释放资源
func (b *CpuBbrQuota) Release() {}

// OnRemoteUpdate 远端更新的时候通知，cpu限流策略是本地模式，不用实现
func (b *CpuBbrQuota) OnRemoteUpdate(_ ratelimiter.RemoteQuotaResult) {

}

// GetQuotaUsed 返回本地限流信息用于上报，cpu限流策略本地模式，不用实现
func (b *CpuBbrQuota) GetQuotaUsed(_ int64) ratelimiter.UsageInfo {
	return ratelimiter.UsageInfo{}
}

// GetAmountInfos 获取规则的限流阈值信息，用于与服务端pb交互
func (b *CpuBbrQuota) GetAmountInfos() []ratelimiter.AmountInfo {
	return nil
}

// createCpuBbrLimiter 初始化一个CPU策略桶
func createCpuBbrLimiter(rule *apitraffic.Rule) *CpuBbrQuota {
	options := make([]bbr.Option, 0)

	if amounts := rule.GetAmounts(); len(amounts) > 0 {
		var amount = amounts[0]
		var minThreshold = amount.GetMaxAmount().GetValue()

		// 有多条规则时，只允许cpu使用率最低的规则生效
		for i := 1; i < len(amounts); i++ {
			if curThreshold := amounts[i].GetMaxAmount().GetValue(); curThreshold < minThreshold {
				minThreshold = curThreshold
				amount = amounts[i]
			}
		}

		// 统计时间窗口，默认 10s
		if window, _ := pb.ConvertDuration(amount.GetValidDuration()); window > 0 {
			options = append(options, bbr.WithWindow(window))
		}
		// 观测时间窗口内 计数桶 的个数（控制滑动窗口精度），默认100个
		if precision := int(amount.GetPrecision().GetValue()); precision > 0 {
			options = append(options, bbr.WithBucket(precision))
		}
		// CPU使用率阈值，默认80%
		if threshold := int64(amount.GetMaxAmount().GetValue()); threshold > 0 {
			options = append(options, bbr.WithCPUThreshold(threshold))
		}
	}

	return &CpuBbrQuota{
		bbr.NewLimiter(options...),
	}
}
