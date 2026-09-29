<template>
  <view class="status-page page-shell">
    <view class="illustration"><view class="paper"><text>—</text><text>——</text><text>—</text></view><text class="leaf">⌁</text><text class="check">✓</text></view>
    <text class="status-title">{{ statusTitle }}</text>
    <text class="status-desc">{{ statusDescription }}</text>

    <view class="timeline">
      <view class="timeline-item done"><text class="dot">✓</text><view><text>私人记录已保存</text><text>内容和亲历信息已进入系统</text></view></view>
      <view class="timeline-item" :class="{ current: status === 'pending_review', done: status === 'published' }"><text class="dot">{{ status === 'published' ? '✓' : '' }}</text><view><text>人工审核</text><text>{{ submittedText }}</text></view></view>
      <view class="timeline-item" :class="{ done: status === 'published' }"><text class="dot">{{ status === 'published' ? '✓' : '' }}</text><view><text>公开展示</text><text>{{ status === 'published' ? '这条体验已经公开' : '审核通过后将公开' }}</text></view></view>
    </view>

    <view class="private-note surface"><text class="book">▢</text><view><text>私人记录已保存</text><text>审核期间，你仍可在足迹中查看和编辑。</text></view></view>
    <button class="primary-button" @tap="goFootprints">查看我的足迹</button>
    <button v-if="status === 'rejected' || status === 'hidden'" class="text-button" @tap="appeal">提交申诉</button>
    <button class="text-button" @tap="goHome">返回发现</button>
    <text class="rule-link">—　了解审核规则 ›　—</text>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
export default {
  data() { return { id: '', status: 'pending_review', submittedAt: '' } },
  computed: {
    statusTitle() { return ({ published: '已经公开', rejected: '需要修改', draft: '私人记录已保存', pending_review: '已提交审核' })[this.status] || '正在处理' },
    statusDescription() { return ({ published: '这条真实到店体验已经可以在发现页看到', rejected: '审核未通过，请回到足迹检查并修改内容', draft: '这条记录目前仅自己可见', pending_review: '你的公开申请已进入人工审核' })[this.status] || '状态更新后会在这里显示' },
    submittedText() { return this.submittedAt ? `${this.formatDate(this.submittedAt)} 提交` : '等待审核进度更新' },
  },
  onLoad(options) { this.id = options.id || ''; this.loadStatus() },
  methods: {
    async loadStatus() {
      if (!this.id) return
      try {
        const result = await api.getVisitRecordReviewStatus(this.id)
        this.status = result.publish_status || this.status
        this.submittedAt = result.submitted_at || ''
      } catch (error) { uni.showToast({ title: error.message || '状态加载失败', icon: 'none' }) }
    },
    formatDate(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '最近' : `${date.getMonth() + 1}月${date.getDate()}日 ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}` },
    goFootprints() { uni.switchTab({ url: '/pages/footprints/footprints' }) },
    goHome() { uni.switchTab({ url: '/pages/home/home' }) },
    appeal() { uni.showModal({ title: '提交申诉', editable: true, placeholderText: '请说明申诉理由（至少 10 个字）', success: async result => { if (!result.confirm) return; try { await api.createAppeal(this.id, result.content || ''); uni.showToast({ title: '申诉已提交', icon: 'success' }) } catch (error) { uni.showToast({ title: error.message || '提交失败', icon: 'none' }) } } }) },
  },
}
</script>

<style scoped>
.status-page { padding-top: 48rpx; text-align: center; }
.illustration { position: relative; width: 240rpx; height: 240rpx; margin: 6rpx auto 34rpx; display: flex; align-items: center; justify-content: center; border: 1rpx solid #C9C3B9; border-radius: 50%; }
.paper { width: 92rpx; height: 126rpx; display: flex; flex-direction: column; justify-content: center; gap: 13rpx; border: 2rpx solid #BEB7AC; border-radius: 8rpx; color: #AAA398; }
.leaf { position: absolute; left: 58rpx; bottom: 44rpx; color: #4F6B57; font-size: 70rpx; transform: rotate(48deg); }
.check { position: absolute; right: 44rpx; bottom: 48rpx; width: 50rpx; height: 50rpx; border-radius: 50%; background: #4F6B57; color: #fff; line-height: 50rpx; }
.status-title { display: block; font-size: 50rpx; font-weight: 700; }
.status-desc { display: block; margin: 18rpx auto 46rpx; color: #7D7871; font-size: 25rpx; line-height: 1.6; }
.timeline { width: 570rpx; margin: 0 auto 46rpx; text-align: left; }
.timeline-item { position: relative; display: flex; gap: 26rpx; min-height: 120rpx; }
.timeline-item:not(:last-child)::after { content: ''; position: absolute; left: 23rpx; top: 50rpx; bottom: -4rpx; border-left: 2rpx dashed #BDB7AF; }
.dot { z-index: 1; width: 48rpx; height: 48rpx; flex: none; border-radius: 50%; background: #D6D2CC; color: #fff; text-align: center; line-height: 48rpx; }
.done .dot { background: #4F6B57; }
.current .dot { border: 13rpx solid #C84B31; background: #fff; }
.timeline-item > view { display: flex; flex-direction: column; }
.timeline-item > view text:first-child { font-size: 30rpx; font-weight: 650; }
.timeline-item > view text:last-child { margin-top: 7rpx; color: #8A857E; font-size: 23rpx; }
.private-note { display: flex; align-items: center; gap: 22rpx; padding: 28rpx; margin-bottom: 36rpx; border-color: #D9A891; text-align: left; }
.book { width: 64rpx; height: 64rpx; border-radius: 50%; background: #F6E7E0; color: #B8442D; line-height: 64rpx; text-align: center; font-size: 32rpx; }
.private-note view { display: flex; flex-direction: column; }
.private-note view text:first-child { font-size: 28rpx; font-weight: 650; }
.private-note view text:last-child { margin-top: 6rpx; color: #817C75; font-size: 22rpx; }
.text-button { margin: 20rpx 0 0; background: transparent; color: #181816; font-size: 27rpx; }
.rule-link { display: block; margin-top: 44rpx; color: #918C85; font-size: 22rpx; }

.status-page { background: #F3F3EF; color: #1B1B19; }
.leaf { color: #426A54; }
.done .dot { background: #426A54; }
.current .dot { border-color: #D64B32; }
.private-note { border-color: #D7D5CF; background: #F8F8F4; }
.book { background: #F4E3DE; color: #D64B32; }
</style>
