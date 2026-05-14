<template>
  <view class="feedback-page">
    <!-- Submit Form -->
    <view class="card form-card" v-if="!showHistory">
      <text class="section-title">提交反馈</text>

      <text class="field-label">反馈类型</text>
      <view class="type-options">
        <text class="type-tag" v-for="t in feedbackTypes" :key="t.value"
          :class="{ active: form.type === t.value }" @tap="form.type = t.value">
          {{ t.label }}
        </text>
      </view>

      <text class="field-label">主题</text>
      <input class="text-input" v-model="form.subject" placeholder="一句话描述问题或建议" :maxlength="50" />

      <text class="field-label">详细内容</text>
      <textarea class="text-area" v-model="form.content" placeholder="请详细描述您遇到的问题或建议..." :maxlength="500" :adjust-position="false" cursor-spacing="20" />

      <text class="field-label">附件（可选，最多3张）</text>
      <view class="attachments">
        <view class="att-item" v-for="(att, idx) in form.attachments" :key="idx">
          <image :src="att.url" class="att-image" mode="aspectFill" />
          <text class="att-remove" @tap="removeAttachment(idx)">✕</text>
        </view>
        <view class="att-add" @tap="addAttachment" v-if="form.attachments.length < 3">
          <text class="att-add-icon">+</text>
        </view>
      </view>

      <button class="btn-primary" @tap="submitFeedback" :loading="submitting">
        {{ submitting ? '提交中...' : '提交反馈' }}
      </button>
    </view>

    <!-- Toggle -->
    <view class="toggle-bar card">
      <text class="toggle-btn" :class="{ active: !showHistory }" @tap="showHistory = false">提交反馈</text>
      <text class="toggle-btn" :class="{ active: showHistory }" @tap="loadHistory">反馈记录</text>
    </view>

    <!-- History List -->
    <view class="history-section" v-if="showHistory">
      <view class="card fb-card" v-for="item in feedbacks" :key="item.id">
        <view class="fb-header">
          <text class="fb-type-tag">{{ typeLabel(item.type) }}</text>
          <text class="fb-status" :class="'status-' + item.status">{{ statusLabel(item.status) }}</text>
        </view>
        <text class="fb-subject">{{ item.subject }}</text>
        <text class="fb-content">{{ item.content }}</text>
        <view class="fb-attachments" v-if="item.attachments && item.attachments.length">
          <image v-for="(att, idx) in item.attachments" :key="idx" :src="att.url" class="fb-att-image" mode="aspectFill" @tap="previewImage(att.url)" />
        </view>
        <view class="fb-reply" v-if="item.reply">
          <text class="reply-label">回复：</text>
          <text class="reply-text">{{ item.reply }}</text>
        </view>
        <text class="fb-time">{{ formatTime(item.created_at) }}</text>
      </view>
      <view class="empty" v-if="feedbacks.length === 0">
        <text class="empty-text">暂无反馈记录</text>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      showHistory: false,
      submitting: false,
      feedbacks: [],
      feedbackTypes: [
        { label: 'Bug 反馈', value: 'bug' },
        { label: '功能建议', value: 'suggestion' },
        { label: '其他', value: 'other' },
      ],
      form: {
        type: 'bug',
        subject: '',
        content: '',
        attachments: [],
      },
    }
  },
  methods: {
    typeLabel(type) {
      const map = { bug: 'Bug', suggestion: '建议', other: '其他' }
      return map[type] || type
    },
    statusLabel(status) {
      const map = { 0: '待处理', 1: '已解决' }
      return map[status] || '未知'
    },
    formatTime(t) {
      if (!t) return ''
      const d = new Date(t)
      return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    addAttachment() {
      uni.chooseImage({
        count: 1,
        sourceType: ['album', 'camera'],
        success: async (res) => {
          const filePath = res.tempFilePaths[0]
          try {
            uni.showLoading({ title: '上传中...' })
            const { object_url } = await api.uploadToCOS(filePath, 'feedback')
            const filename = filePath.split('/').pop()
            this.form.attachments.push({ name: filename, url: object_url })
          } catch (err) {
            uni.showToast({ title: '上传失败', icon: 'none' })
          } finally {
            uni.hideLoading()
          }
        },
      })
    },
    removeAttachment(idx) {
      this.form.attachments.splice(idx, 1)
    },
    async submitFeedback() {
      if (!this.form.subject.trim()) {
        uni.showToast({ title: '请填写主题', icon: 'none' })
        return
      }
      if (!this.form.content.trim()) {
        uni.showToast({ title: '请填写内容', icon: 'none' })
        return
      }
      this.submitting = true
      try {
        await api.createFeedback(this.form)
        uni.showToast({ title: '提交成功，感谢反馈', icon: 'success' })
        this.form = { type: 'bug', subject: '', content: '', attachments: [] }
      } catch (err) {
        uni.showToast({ title: '提交失败: ' + (err.message || ''), icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    async loadHistory() {
      this.showHistory = true
      try {
        const res = await api.getFeedbacks()
        this.feedbacks = res.feedbacks || []
      } catch (err) {
        console.log('load feedbacks failed', err)
      }
    },
    previewImage(url) {
      uni.previewImage({ urls: [url], current: url })
    },
  },
}
</script>

<style scoped>
.feedback-page {
  padding: 24rpx;
  min-height: 100vh;
  background: #f5f5f5;
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 34rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 28rpx;
}

.field-label {
  font-size: 26rpx;
  color: #888;
  display: block;
  margin-bottom: 12rpx;
  margin-top: 24rpx;
}

.type-options {
  display: flex;
  gap: 16rpx;
  margin-bottom: 8rpx;
}

.type-tag {
  padding: 12rpx 28rpx;
  border-radius: 30rpx;
  background: #f5f5f5;
  font-size: 26rpx;
  color: #666;
}

.type-tag.active {
  background: #e8f5e9;
  color: #4CAF50;
  font-weight: bold;
}

.text-input {
  background: #f5f5f5;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  width: 100%;
  color: #333;
}

.text-area {
  background: #f5f5f5;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  width: 100%;
  min-height: 200rpx;
  color: #333;
}

/* Attachments */
.attachments {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.att-item {
  position: relative;
}

.att-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 12rpx;
}

.att-remove {
  position: absolute;
  top: -10rpx;
  right: -10rpx;
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border-radius: 50%;
  font-size: 24rpx;
}

.att-add {
  width: 140rpx;
  height: 140rpx;
  border: 2rpx dashed #ccc;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.att-add-icon {
  font-size: 52rpx;
  color: #ccc;
}

/* Button */
.btn-primary {
  width: 100%;
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border: none;
  border-radius: 50rpx;
  padding: 24rpx 0;
  font-size: 30rpx;
  margin-top: 16rpx;
}

.btn-primary::after {
  border: none;
}

/* Toggle */
.toggle-bar {
  display: flex;
  overflow: hidden;
  padding: 0 !important;
}

.toggle-btn {
  flex: 1;
  text-align: center;
  padding: 24rpx;
  font-size: 28rpx;
  color: #999;
}

.toggle-btn.active {
  color: #4CAF50;
  font-weight: bold;
  border-bottom: 4rpx solid #4CAF50;
}

/* Feedback Card */
.fb-card {
  margin-bottom: 20rpx;
}

.fb-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.fb-type-tag {
  font-size: 22rpx;
  background: #e8f5e9;
  color: #4CAF50;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.fb-status {
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.status-0 {
  background: #fff3e0;
  color: #FF9800;
}

.status-1 {
  background: #e8f5e9;
  color: #4CAF50;
}

.fb-subject {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.fb-content {
  font-size: 26rpx;
  color: #666;
  display: block;
  line-height: 1.7;
  margin-bottom: 12rpx;
}

.fb-attachments {
  display: flex;
  gap: 12rpx;
  margin-bottom: 12rpx;
}

.fb-att-image {
  width: 120rpx;
  height: 120rpx;
  border-radius: 8rpx;
}

.fb-reply {
  background: #f0f7ff;
  border-radius: 8rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 12rpx;
}

.reply-label {
  font-size: 24rpx;
  color: #1976D2;
  font-weight: bold;
  display: block;
  margin-bottom: 6rpx;
}

.reply-text {
  font-size: 26rpx;
  color: #333;
  line-height: 1.6;
}

.fb-time {
  font-size: 22rpx;
  color: #bbb;
  display: block;
  margin-top: 8rpx;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
}

.empty-text {
  font-size: 28rpx;
  color: #ccc;
}
</style>
