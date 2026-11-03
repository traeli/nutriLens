<template>
  <view class="record-page">
    <!-- 入口 hub -->
    <view v-if="!recordMode" class="entry-page" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="entry-content">
        <view class="entry-heading">
          <view class="entry-title-row">
            <view class="entry-title-ic" :style="dz('pen', '#3E9B6F', 38)"></view>
            <text class="entry-title">新建记录</text>
          </view>
          <text class="entry-sub">先选择记录类型，两条流程会分别保存草稿</text>
        </view>

        <view class="nutrition-entry" @tap="openNutrition">
          <view class="nut-medal" :style="dzBg('blob-white')"><view class="nut-ic" :style="dz('mifan', '#332E26', 40)"></view></view>
          <view class="nut-copy">
            <view class="nut-title-row"><text class="card-num">02</text><text class="nut-title">识别餐食营养</text></view>
            <text class="nut-note">拍菜品估算营养，不要拍账单 · 仅自己可见</text>
          </view>
          <text class="nut-arrow">›</text>
        </view>

        <view class="visit-entry">
          <view class="visit-cover">
            <text class="cover-sticker">好好吃饭</text>
            <view class="cover-burst" :style="dz('burst', '#FFC64B', 56)"></view>
            <view class="cover-stop" :style="dzBg('blob-white')"><view class="stop-ic" :style="dz('bowl', '#332E26', 50)"></view></view>
            <view class="cover-squig" :style="dz('squiggle', '#D8CCB0', 38)"></view>
            <view class="cover-stop two" :style="dzBg('blob-white')"><view class="stop-ic" :style="dz('chuaner', '#332E26', 50)"></view></view>
            <view class="cover-squig" :style="dz('squiggle', '#D8CCB0', 38)"></view>
            <view class="cover-stop three" :style="dzBg('blob-white')"><view class="stop-ic" :style="dz('coffee', '#332E26', 50)"></view></view>
          </view>
          <view class="visit-entry-head">
            <view class="visit-name-row"><text class="card-num">01</text><text class="visit-name">到店体验</text></view>
            <text class="visit-city">{{ city }}</text>
          </view>

          <view class="direct-entry" @tap="startVisit('manual')">
            <view class="direct-ic" :style="dz('noodle', '#FFFFFF', 56)"></view>
            <view class="direct-copy">
              <text class="direct-title">记录一次到店体验</text>
              <text class="direct-note">选店铺、写结论，再补照片或消费凭证</text>
            </view>
            <view class="direct-arrow" :style="dz('arrow', '#FFFFFF', 28)"></view>
          </view>

          <view class="capture-grid">
            <view class="capture-entry" @tap="openVoicePanel">
              <view class="capture-medal" :style="dzBg('blob-white')"><view class="capture-ic" :style="dz('quote', '#332E26', 34)"></view></view>
              <text class="capture-title">说说体验</text>
              <text class="capture-note">语音整理后再确认</text>
            </view>
            <view class="capture-entry" @tap="captureReceipt">
              <view class="capture-medal flip" :style="dzBg('blob-white')"><view class="capture-ic" :style="dz('stamp', '#332E26', 34)"></view></view>
              <text class="capture-title">拍消费单</text>
              <text class="capture-note">识别店名与消费信息</text>
            </view>
          </view>
        </view>

        <view class="recent-submissions">
          <view class="sec">
            <view class="sec-ic" :style="dz('check', '#3E9B6F', 34)"></view>
            <view class="sec-title">
              <text class="sec-h">最近评论</text>
              <view class="sec-scribble" :style="dz('scribble', '#3E9B6F', 52)"></view>
            </view>
            <text class="sec-p">提交后马上能看审核进度</text>
            <text class="sec-link" @tap="openMyRecords">全部 ›</text>
          </view>
          <view class="recent-card">
            <view v-if="recentLoading" class="recent-empty">正在整理评论…</view>
            <view v-else-if="!recentRecords.length" class="recent-empty">
              <view class="empty-trail">
                <view class="trail-stop" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('bowl', '#332E26', 38)"></view></view>
                <view class="trail-squig" :style="dz('squiggle', '#D8CCB0', 34)"></view>
                <view class="trail-stop flip" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('noodle', '#332E26', 38)"></view></view>
                <view class="trail-squig" :style="dz('squiggle', '#D8CCB0', 34)"></view>
                <view class="trail-stop" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('coffee', '#332E26', 38)"></view></view>
              </view>
              <text class="empty-note">还没有餐厅评论，认真记下第一餐吧。</text>
            </view>
            <view v-for="item in recentRecords" :key="item.id" class="recent-row" @tap="openSubmittedRecord(item)">
              <view class="recent-mark">{{ item.place.slice(0, 1) }}</view>
              <view class="recent-copy">
                <text class="recent-place">{{ item.place }}</text>
                <text class="recent-meta">{{ item.date }} · {{ item.conclusion }}</text>
              </view>
              <text class="record-status" :class="`status-${item.status}`">{{ item.statusLabel }}</text>
              <text class="recent-arrow">›</text>
            </view>
          </view>
        </view>
      </view>

      <view v-if="voicePanel" class="capture-mask" @tap="closeVoicePanel">
        <view class="voice-panel" @tap.stop>
          <view class="panel-handle"></view>
          <text class="voice-panel-title">{{ recording ? '正在听你说' : '说说这次体验' }}</text>
          <view class="voice-orbit" :class="{ recording }" @tap="toggleRecording">
            <view class="voice-core"><view></view></view>
          </view>
          <text class="voice-time">{{ recording ? formatSeconds(recordSeconds) : '点击开始录音' }}</text>
          <text class="voice-cancel" @tap="closeVoicePanel">取消</text>
        </view>
      </view>
    </view>

    <!-- 评分表单 -->
    <view v-else class="rating-page">
      <view class="rating-nav" :style="{ paddingTop: `${navLayout.contentTop}px` }">
        <view class="mode-back" @tap="backToModes"><text class="back-mark">‹</text><text>返回</text></view>
        <text class="rating-nav-title">记一餐</text>
        <view class="switch-method" @tap="backToModes">换个方式</view>
      </view>

      <scroll-view class="rating-scroll" scroll-y scroll-with-animation :scroll-into-view="formScrollTarget">
        <view class="rating-canvas">
          <view id="restaurantField" class="field-card" :class="{ 'field-error': placeError }">
            <view class="tape"></view>
            <view class="field-topline">
              <view class="field-name-row"><text class="card-num">01</text><text class="field-name">到店记录</text></view>
              <text class="field-city">{{ city }}</text>
            </view>
            <text class="writing-label">这次去了哪家店？</text>
            <input
              class="place-input"
              v-model="place"
              :focus="restaurantInputFocus"
              maxlength="160"
              placeholder="输入餐厅名称"
              placeholder-class="writing-placeholder"
              confirm-type="done"
              @input="onPlaceInput"
              @focus="onPlaceFocus"
            />
            <view class="input-foot">
              <text class="foot-hint" :class="{ error: placeError }">{{ placeError || placeSelectionHint }}</text>
              <text class="foot-counter">{{ place.length }}/160</text>
            </view>

            <view v-if="showPlaceSuggestions" class="place-suggestions">
              <view v-if="placeSearching" class="suggestion-state">正在搜索餐厅…</view>
              <view v-for="item in placeSuggestions" :key="item.id" class="suggestion-row" @tap="choosePlaceSuggestion(item)">
                <view class="suggestion-copy">
                  <text class="suggestion-name">{{ item.name }}</text>
                  <text class="suggestion-meta">{{ placeSuggestionMeta(item) }}</text>
                </view>
                <text class="suggestion-pick">选择</text>
              </view>
              <view v-if="place.trim() && !placeSearching" class="suggestion-create" @tap="choosePlaceOnMap">
                <view class="create-ic" :style="dz('plus', '#2F7D58', 26)"></view>
                <view class="create-copy">
                  <text class="create-title">在微信地图选择这家店</text>
                  <text class="create-note">自动获取准确地址和坐标</text>
                </view>
              </view>
            </view>

            <view v-if="selectedPlaceLocation" class="selected-place-location">
              <view class="selected-copy">
                <text class="selected-name">{{ selectedPlaceLocation.name }}</text>
                <text class="selected-meta">{{ selectedPlaceLocation.address }}</text>
              </view>
              <text class="selected-repick" @tap="choosePlaceOnMap">重选</text>
            </view>
            <button v-else-if="!placeId" class="choose-location-button" @tap="choosePlaceOnMap">
              <view class="locate-ic" :style="dz('pin', '#3E9B6F', 26)"></view>
              <text>从微信地图选择餐厅位置</text>
            </button>
          </view>

          <view id="feelingField" class="field-card" :class="{ 'field-error': conclusionError }">
            <view class="field-topline">
              <view class="field-name-row"><text class="card-num">02</text><text class="field-name">这一顿，感觉如何？</text></view>
            </view>
            <view class="feeling-picker" :class="{ selected: selectedConclusion }" @tap="openFeelingSheet">
              <view v-if="selectedConclusion" class="feeling-selected">
                <view class="feeling-dot" :class="selectedConclusion.value"></view>
                <view class="feeling-copy">
                  <text class="feeling-label">{{ selectedConclusion.label }}</text>
                  <text class="feeling-desc">{{ selectedConclusion.desc }}</text>
                </view>
              </view>
              <view v-else class="feeling-placeholder">
                <text class="feeling-label placeholder">选择本次感受</text>
                <text class="feeling-desc">{{ conclusionError || '请选择最接近真实体验的一项' }}</text>
              </view>
              <view class="picker-caret" :style="dz('caret', '#8B8271', 20)"></view>
            </view>
          </view>

          <view class="story-entry" :class="{ expanded: detailsExpanded }">
            <view class="story-trigger" @tap="detailsExpanded = !detailsExpanded">
              <view class="story-trigger-copy">
                <text class="story-trigger-title">再多记一点（选填）</text>
                <text class="story-trigger-note">简单记两句，或直接提交审核</text>
              </view>
              <text class="story-toggle">{{ detailsExpanded ? '−' : '＋' }}</text>
            </view>

            <view v-if="detailsExpanded" class="story-content">
              <textarea v-model="content" maxlength="500" :placeholder="experiencePlaceholder" />

              <view class="record-tags">
                <text class="tags-label">体验标签（选填）</text>
                <view class="tags-row">
                  <text v-for="tag in availableTags" :key="tag.code" class="tag-chip" :class="{ active: selectedTags.includes(tag.code) }" @tap="toggleTag(tag.code)">{{ tag.name }}</text>
                </view>
              </view>

              <view class="detail-tools">
                <text class="tool-chip" :class="{ active: showCost }" @tap="showCost = !showCost">＋ 人均消费</text>
                <text class="tool-chip" :class="{ active: showWait }" @tap="showWait = !showWait">＋ 排队时间</text>
                <text class="tool-chip" :class="{ active: showDish }" @tap="showDish = !showDish">＋ 推荐菜</text>
                <text class="tool-chip" @tap="chooseExperiencePhotos">＋ 添加照片</text>
              </view>

              <view v-if="showCost || showWait || showDish" class="detail-fields">
                <view v-if="showCost" class="detail-field"><text class="detail-field-label">人均</text><input class="detail-field-input" v-model="averageCost" type="digit" placeholder="0" /><text class="detail-field-unit">元</text></view>
                <view v-if="showWait" class="detail-field"><text class="detail-field-label">排队</text><input class="detail-field-input" v-model="waitMinutes" type="number" placeholder="0" /><text class="detail-field-unit">分钟</text></view>
                <view v-if="showDish" class="detail-field"><text class="detail-field-label">推荐菜</text><input class="detail-field-input detail-field-text" v-model="recommendedDish" placeholder="写下菜名" /></view>
              </view>

              <scroll-view v-if="allPhotos.length" class="photo-scroll" scroll-x>
                <view class="photo-row">
                  <image v-for="photo in allPhotos" :key="photo" :src="photo" mode="aspectFill" />
                </view>
              </scroll-view>
            </view>
          </view>

          <text v-if="!canRecord" class="submit-hint">{{ missingPrompt }}</text>
          <text class="review-hint">提交后进入后台审核，通过后才会公开</text>
          <button class="stamp-button" :class="{ ready: canRecord }" :disabled="submitting" :loading="submitting" @tap="finishRecord">
            <text>{{ submitting ? '正在提交' : '盖个章，完成记录' }}</text>
            <view class="stamp-arrow" :style="dz('arrow', canRecord ? '#332E26' : '#8B8271', 28)"></view>
          </button>
        </view>
      </scroll-view>

      <view v-if="feelingSheet" class="feeling-mask" @tap="closeFeelingSheet">
        <view class="feeling-sheet" @tap.stop>
          <view class="panel-handle"></view>
          <text class="feeling-sheet-title">这一顿，感觉如何？</text>
          <text class="feeling-sheet-desc">选择最接近真实体验的一项</text>
          <view class="feeling-sheet-options">
            <view v-for="item in conclusions" :key="item.value" class="feeling-sheet-option" :class="{ active: conclusion === item.value }" @tap="selectConclusion(item.value)">
              <view class="feeling-dot" :class="item.value"></view>
              <view class="feeling-copy">
                <text class="feeling-label">{{ item.label }}</text>
                <text class="feeling-desc">{{ item.desc }}</text>
              </view>
              <view v-if="conclusion === item.value" class="option-check" :style="dz('check', '#2F7D58', 28)"></view>
              <text v-else class="option-arrow">›</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { assetUrl } from '@/utils/assets.js'
import { diningImages } from '@/mock/city-dining.js'
import { getSelectedCity, refreshCityOptions } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'
import { doodleBg, doodleStyle } from '@/utils/doodles.js'

export default {
  data() {
    const selectedCity = getSelectedCity()
    return {
      navLayout: getPrimaryNavLayout(),
      entryArt: assetUrl('/static/dining/visit-rating-art-v2.jpg'),
      recordMode: '',
      entrySource: 'manual',
      images: diningImages,
      city: selectedCity.name,
      cityCode: selectedCity.code,
      place: '',
      placeId: 0,
      selectedPlaceLocation: null,
      placeSuggestions: [],
      placeSearching: false,
      showPlaceSuggestions: false,
      placeSearchTimer: null,
      placeError: '',
      restaurantInputFocus: false,
      conclusion: '',
      conclusions: [
        { value: 'hot', label: '夯', desc: '值得专程再来' },
        { value: 'recommend', label: '推荐', desc: '整体体验不错' },
        { value: 'neutral', label: '一般', desc: '没有明显惊喜' },
        { value: 'caution', label: '谨慎', desc: '有些问题需要留意' },
      ],
      feelingSheet: false,
      conclusionError: '',
      formScrollTarget: '',
      detailsExpanded: false,
      content: '',
      showCost: false,
      showWait: false,
      showDish: false,
      averageCost: '',
      waitMinutes: '',
      recommendedDish: '',
      availableTags: [{ code: 'taste', name: '口味' }, { code: 'price', name: '价格' }, { code: 'service', name: '服务' }, { code: 'queue', name: '排队' }, { code: 'hygiene_observation', name: '卫生观感' }],
      selectedTags: [],
      receiptImage: '',
      experiencePhotos: [],
      voicePanel: false,
      recording: false,
      recordSeconds: 0,
      recordingTimer: null,
      recorderManager: null,
      discardRecording: false,
      transcribing: false,
      submitting: false,
      recordRequestKey: '',
      recentLoading: false,
      recentRecords: [],
    }
  },
  computed: {
    canRecord() {
      return Boolean(this.place.trim() && (this.placeId || this.selectedPlaceLocation) && this.conclusion)
    },
    selectedConclusion() {
      return this.conclusions.find(item => item.value === this.conclusion) || null
    },
    placeSelectionHint() {
      if (this.placeId) return '已选择收录餐厅'
      if (this.selectedPlaceLocation) return '已从微信地图确认位置'
      return '新店需从微信地图确认位置'
    },
    missingPrompt() {
      const missing = []
      if (!this.place.trim()) missing.push('餐厅')
      else if (!this.placeId && !this.selectedPlaceLocation) missing.push('餐厅位置')
      if (!this.conclusion) missing.push('用餐感受')
      return missing.length ? `还差：${missing.join('、')}` : ''
    },
    allPhotos() {
      return [this.receiptImage, ...this.experiencePhotos].filter(Boolean)
    },
    experiencePlaceholder() {
      return ({
        hot: '最值得专程来吃的是什么？',
        recommend: '哪一点让你愿意推荐？',
        neutral: '哪些地方还可以更好？',
        caution: '有什么需要其他人留意？',
      })[this.conclusion] || '写下真实的体验感受…'
    },
  },
  onLoad() {
    this.setupRecorder()
  },
  onShow() {
    this.navLayout = getPrimaryNavLayout()
    syncCustomTabBar(this, 1)
    const selectedCity = getSelectedCity()
    this.city = selectedCity.name
    this.cityCode = selectedCity.code
    this.loadRecentRecords()
  },
  onUnload() {
    this.stopRecordingTimer()
    if (this.placeSearchTimer) clearTimeout(this.placeSearchTimer)
    if (this.recording && this.recorderManager) {
      this.discardRecording = true
      this.recorderManager.stop()
    }
  },
  methods: {
    dz(key, color, size) { return doodleStyle(key, color, size) },
    dzBg(key) { return doodleBg(key) },
    setupRecorder() {
      // #ifdef MP-WEIXIN
      this.recorderManager = uni.getRecorderManager()
      this.recorderManager.onStart(() => {
        this.recording = true
        this.recordSeconds = 0
        this.stopRecordingTimer()
        this.recordingTimer = setInterval(() => { this.recordSeconds += 1 }, 1000)
      })
      this.recorderManager.onStop(result => {
        this.recording = false
        this.stopRecordingTimer()
        if (this.discardRecording) {
          this.discardRecording = false
          return
        }
        if (!result || !result.tempFilePath) return
        this.voicePanel = false
        this.transcribeRecording(result.tempFilePath)
      })
      this.recorderManager.onError(() => {
        this.recording = false
        this.stopRecordingTimer()
        uni.showToast({ title: '录音失败，请重试', icon: 'none' })
      })
      // #endif
    },
    stopRecordingTimer() {
      if (this.recordingTimer) clearInterval(this.recordingTimer)
      this.recordingTimer = null
    },
    startVisit(source = 'manual') {
      this.entrySource = source
      this.detailsExpanded = source !== 'manual'
      if (!this.recordRequestKey) this.recordRequestKey = `visit-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
      this.recordMode = 'visit'
    },
    backToModes() {
      this.recordMode = ''
    },
    openVoicePanel() {
      if (!this.recorderManager) {
        uni.showToast({ title: '请在微信小程序中使用语音记录', icon: 'none' })
        return
      }
      this.discardRecording = false
      this.voicePanel = true
    },
    closeVoicePanel() {
      if (this.recording && this.recorderManager) {
        this.discardRecording = true
        this.recorderManager.stop()
      }
      this.voicePanel = false
    },
    toggleRecording() {
      if (!this.recorderManager) return
      if (this.recording) {
        this.recorderManager.stop()
        return
      }
      this.discardRecording = false
      this.recorderManager.start({ duration: 60000, sampleRate: 16000, numberOfChannels: 1, encodeBitRate: 48000, format: 'mp3' })
    },
    formatSeconds(value) {
      const minutes = String(Math.floor(value / 60)).padStart(2, '0')
      const seconds = String(value % 60).padStart(2, '0')
      return `${minutes}:${seconds}`
    },
    async transcribeRecording(filePath) {
      if (this.transcribing) return
      this.transcribing = true
      uni.showLoading({ title: '正在整理语音', mask: true })
      try {
        const result = await api.transcribeVisitAudio(filePath)
        this.content = String((result && result.text) || '').trim()
        this.startVisit('voice')
      } catch (error) {
        this.startVisit('voice')
        uni.showToast({ title: error.message || '语音识别失败，可手动填写', icon: 'none' })
      } finally {
        this.transcribing = false
        uni.hideLoading()
      }
    },
    captureReceipt() {
      uni.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['camera', 'album'],
        success: result => {
          const file = result && result.tempFiles && result.tempFiles[0]
          if (!file || !file.tempFilePath) return
          this.receiptImage = file.tempFilePath
          this.startVisit('receipt')
        },
      })
    },
    chooseExperiencePhotos() {
      uni.chooseMedia({
        count: 3,
        mediaType: ['image'],
        sourceType: ['camera', 'album'],
        success: result => {
          const files = result && Array.isArray(result.tempFiles) ? result.tempFiles : []
          this.experiencePhotos = files.map(item => item.tempFilePath).filter(Boolean)
        },
      })
    },
    onPlaceInput(event) {
      this.place = String((event && event.detail && event.detail.value) || this.place)
      this.placeId = 0
      this.selectedPlaceLocation = null
      this.placeError = ''
      if (this.placeSearchTimer) clearTimeout(this.placeSearchTimer)
      const keyword = this.place.trim()
      if (!keyword) {
        this.placeSuggestions = []
        this.showPlaceSuggestions = false
        return
      }
      this.showPlaceSuggestions = true
      this.placeSearchTimer = setTimeout(() => this.searchPlaceSuggestions(keyword), 300)
    },
    onPlaceFocus() {
      if (this.place.trim() && !this.placeId) this.showPlaceSuggestions = true
    },
    async searchPlaceSuggestions(keyword) {
      if (keyword !== this.place.trim()) return
      this.placeSearching = true
      try {
        const page = await api.searchPlaces({ city_code: this.cityCode, q: keyword, limit: 6 })
        if (keyword !== this.place.trim()) return
        this.placeSuggestions = page.items || []
      } catch (error) {
        this.placeSuggestions = []
        console.warn('[record] place suggestions load failed:', error.message)
      } finally {
        if (keyword === this.place.trim()) this.placeSearching = false
      }
    },
    choosePlaceSuggestion(item) {
      this.placeId = Number(item.id || 0)
      this.place = item.name || this.place
      this.selectedPlaceLocation = item.longitude && item.latitude ? {
        name: item.name || this.place,
        address: item.address || '',
        longitude: Number(item.longitude),
        latitude: Number(item.latitude),
      } : null
      this.placeError = ''
      this.showPlaceSuggestions = false
      this.placeSuggestions = []
      this.restaurantInputFocus = false
      if (typeof uni.vibrateShort === 'function') uni.vibrateShort({ type: 'light' })
    },
    choosePlaceOnMap() {
      uni.chooseLocation({
        success: result => {
          const longitude = Number(result.longitude)
          const latitude = Number(result.latitude)
          if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) {
            uni.showToast({ title: '没有获得有效坐标，请重试', icon: 'none' })
            return
          }
          const address = String(result.address || '').trim()
          if (!address) {
            uni.showToast({ title: '请选择带详细地址的门店位置', icon: 'none' })
            return
          }
          this.placeId = 0
          this.place = String(result.name || this.place).trim()
          this.selectedPlaceLocation = {
            name: this.place,
            address,
            longitude,
            latitude,
          }
          this.placeError = ''
          this.showPlaceSuggestions = false
          this.placeSuggestions = []
          this.restaurantInputFocus = false
        },
        fail: error => {
          const message = String((error && error.errMsg) || '')
          if (!message.includes('cancel')) uni.showToast({ title: '请允许位置权限后选择餐厅', icon: 'none' })
        },
      })
    },
    placeSuggestionMeta(item) {
      return [item.business_area || item.district, item.address].filter(Boolean).join(' · ') || '地址待补充'
    },
    openFeelingSheet() { this.feelingSheet = true },
    closeFeelingSheet() { this.feelingSheet = false },
    selectConclusion(value) {
      this.conclusion = value
      this.conclusionError = ''
      this.feelingSheet = false
    },
    toggleTag(code) { this.selectedTags = this.selectedTags.includes(code) ? this.selectedTags.filter(item => item !== code) : [...this.selectedTags, code] },
    async ensureRecordCity() {
      let selectedCity = getSelectedCity()
      if (!selectedCity.code) {
        await refreshCityOptions()
        selectedCity = getSelectedCity()
      }
      if (!String(selectedCity.code || '').trim()) throw new Error('暂无可用城市，请回首页选择城市后重试')
      this.city = selectedCity.name
      this.cityCode = String(selectedCity.code).trim()
    },
    async finishRecord() {
      if (this.submitting || !this.validateRequiredFields()) return
      if (!uni.getStorageSync('token')) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      this.submitting = true
      let created = null
      try {
        await this.ensureRecordCity()
        let placeId = this.placeId
        if (!placeId) {
          const location = this.selectedPlaceLocation
          const submittedPlace = await api.createPlace({
            name: this.place.trim(), city_code: this.cityCode, category: 'restaurant',
            address: location.address, longitude: location.longitude, latitude: location.latitude,
            poi_provider: 'wechat',
          })
          placeId = Number(submittedPlace.id || 0)
          if (!placeId) throw new Error('地点创建失败，请重试')
        }
        const dishes = this.recommendedDish.trim() ? [{ name: this.recommendedDish.trim() }] : []
        created = await api.createVisitRecord({
          place_id: placeId, city_code: this.cityCode, visit_date: this.today(),
          consumer_type: 'self', conclusion: this.conclusion,
          average_cost: this.averageCost === '' ? null : Number(this.averageCost),
          wait_minutes: this.waitMinutes === '' ? null : Number(this.waitMinutes),
          dishes, content: this.content.trim(), visibility: 'private', tag_codes: this.selectedTags,
        }, this.recordRequestKey)
        if (this.receiptImage) await api.uploadEvidence(created.record.id, this.receiptImage, 'receipt')
        for (const photo of this.experiencePhotos) await api.uploadRecordMedia(created.record.id, photo, 'photo')
        const submitted = await api.createRestaurantReview(created.record.id)
        this.resetVisitForm()
        this.recordMode = ''
        await this.loadRecentRecords()
        uni.showModal({
          title: '提交成功',
          content: '到店足迹已经保存，餐厅评论已提交审核。你可以在本页下方或“我的评论”查看进度。',
          showCancel: false,
          confirmText: '知道了',
        })
        return submitted
      } catch (error) {
        uni.showToast({ title: created ? '已保存草稿，提交审核失败' : (error.message || '提交失败，请重试'), icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    validateRequiredFields() {
      this.placeError = !this.place.trim() ? '请先填写或选择餐厅' : (!this.placeId && !this.selectedPlaceLocation ? '请从微信地图确认餐厅位置' : '')
      this.conclusionError = this.conclusion ? '' : '请选择本次用餐感受'
      if (!this.placeError && !this.conclusionError) return true
      if (typeof uni.vibrateShort === 'function') uni.vibrateShort({ type: 'medium' })
      const target = this.placeError ? 'restaurantField' : 'feelingField'
      this.formScrollTarget = ''
      this.$nextTick(() => {
        this.formScrollTarget = target
        if (this.placeError) this.restaurantInputFocus = true
        if (!this.placeError && this.conclusionError) this.feelingSheet = true
      })
      uni.showToast({ title: this.placeError || this.conclusionError, icon: 'none' })
      return false
    },
    async loadRecentRecords() {
      if (!uni.getStorageSync('token')) { this.recentRecords = []; return }
      this.recentLoading = true
      try {
        const page = await api.getMyReviews({ limit: 10 })
        this.recentRecords = (page.items || []).slice(0, 3).map(this.normalizeSubmittedRecord)
      } catch (error) {
        console.warn('[record] recent submissions load failed:', error.message)
      } finally { this.recentLoading = false }
    },
    normalizeSubmittedRecord(item) {
      const status = item.review.publish_status
      const labels = { pending: '待审核', published: '审核通过', rejected: '审核不通过' }
      const conclusions = { hot: '夯', recommend: '推荐', neutral: '一般', caution: '谨慎' }
      return {
        id: item.review.id, status, statusLabel: labels[status] || '处理中',
        place: item.place.name || '未命名餐厅', conclusion: conclusions[item.version.conclusion] || '已记录',
        date: this.formatVisitDate(item.version.visit_date),
      }
    },
    formatVisitDate(value) {
      const text = String(value || '').slice(0, 10)
      return text ? text.replace(/^\d{4}-/, '').replace('-', '月') + '日' : '刚刚'
    },
    today() {
      const date = new Date()
      const pad = value => String(value).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
    },
    resetVisitForm() {
      this.place = ''; this.placeId = 0; this.selectedPlaceLocation = null; this.placeSuggestions = []; this.showPlaceSuggestions = false; this.placeError = ''; this.restaurantInputFocus = false
      this.conclusion = ''; this.conclusionError = ''; this.content = ''; this.selectedTags = []; this.averageCost = ''; this.waitMinutes = ''; this.recommendedDish = ''
      this.detailsExpanded = false; this.showCost = false; this.showWait = false; this.showDish = false
      this.receiptImage = ''; this.experiencePhotos = []; this.recordRequestKey = ''
    },
    openSubmittedRecord(item) {
      const url = item.status === 'published' ? `/pages/experience-detail/experience-detail?id=${item.id}` : `/pages/review-status/review-status?id=${item.id}`
      uni.navigateTo({ url })
    },
    openMyRecords() {
      uni.navigateTo({ url: '/pages/my-records/my-records' })
    },
    openNutrition() { uni.navigateTo({ url: '/pages/nutrition/nutrition' }) },
  },
}
</script>

<style scoped>
/* —— V6 贴纸手帐：token 同 home.vue —— */
.record-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #FFFDF9 0%, #FFFBF3 36%, #FFF3DF 100%);
  color: #332E26;
  font-family: "PingFang SC", "Helvetica Neue", Helvetica, sans-serif;
}

/* ===== 入口 hub ===== */
.entry-page { padding: 0 32rpx calc(108rpx + env(safe-area-inset-bottom)); overflow: hidden; }
.entry-content { display: flex; flex-direction: column; }
.entry-heading { order: 1; padding: 8rpx 4rpx 0; }
.entry-title-row { display: flex; align-items: center; gap: 14rpx; }
.entry-title-ic { flex: none; }
.entry-title { font-size: 40rpx; font-weight: 900; letter-spacing: 2rpx; line-height: 1; }
.entry-sub { display: block; margin-top: 12rpx; font-size: 24rpx; color: #8B8271; }
.card-num { color: #3E9B6F; font-size: 22rpx; font-weight: 800; letter-spacing: 2rpx; }

/* 营养速记（奶油卡，次入口） */
.nutrition-entry {
  order: 3;
  display: flex; align-items: center; gap: 20rpx; margin-top: 28rpx;
  padding: 24rpx 28rpx; border-radius: 36rpx;
  background: #FFF4DF; border: 3rpx solid #EADFC2;
}
.nut-medal {
  flex: none; width: 92rpx; height: 92rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center; transform: rotate(-4deg);
}
.nut-ic { flex: none; }
.nut-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.nut-title-row { display: flex; align-items: center; gap: 10rpx; }
.nut-title { font-size: 28rpx; font-weight: 800; }
.nut-note { font-size: 21rpx; color: #8B8271; }
.nut-arrow { flex: none; font-size: 32rpx; font-weight: 700; color: #C9BFA8; }

/* 到店体验大卡 */
.visit-entry {
  order: 2;
  margin-top: 24rpx; padding: 28rpx 28rpx 32rpx;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 44rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.recent-submissions { order: 4; }
.visit-cover {
  position: relative; display: flex; align-items: center; justify-content: center; gap: 14rpx;
  height: 176rpx; border-radius: 30rpx; box-sizing: border-box;
  background: #FFF4DF; border: 3rpx solid #EADFC2;
}
.cover-stop {
  width: 108rpx; height: 108rpx; box-sizing: border-box; transform: rotate(-6deg);
  display: flex; align-items: center; justify-content: center;
}
.cover-stop.two { transform: rotate(4deg) scaleX(-1); }
.cover-stop.two .stop-ic { transform: scaleX(-1); }
.cover-stop.three { transform: rotate(-3deg); }
.stop-ic { flex: none; }
.cover-squig { flex: none; }
.cover-burst {
  position: absolute; left: -8rpx; bottom: -12rpx; transform: rotate(-12deg); opacity: .95;
}
.cover-sticker {
  position: absolute; top: -14rpx; right: 22rpx; transform: rotate(3deg);
  padding: 6rpx 16rpx; border-radius: 14rpx;
  background: linear-gradient(180deg, #FFD05C 0%, #FFC64B 60%, #F7B93E 100%); box-shadow: 0 3rpx 0 #E0A62F;
  font-family: "Kaiti SC", "STKaiti", "PingFang SC", sans-serif; font-size: 22rpx; letter-spacing: 2rpx;
}
.visit-entry-head { display: flex; align-items: center; justify-content: space-between; margin-top: 24rpx; }
.visit-name-row { display: flex; align-items: center; gap: 10rpx; }
.visit-name { font-size: 30rpx; font-weight: 900; }
.visit-city { font-size: 22rpx; color: #8B8271; }
.direct-entry {
  display: flex; align-items: center; gap: 18rpx; margin-top: 20rpx;
  height: 108rpx; padding: 0 30rpx; border-radius: 999rpx;
  background: linear-gradient(180deg, #4BAB7E 0%, #3E9B6F 52%, #35855A 100%); box-shadow: 0 6rpx 0 #2F7D58;
}
.direct-ic { flex: none; }
.direct-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.direct-title { color: #FFFFFF; font-size: 30rpx; font-weight: 800; }
.direct-note { color: rgba(255, 255, 255, .82); font-size: 21rpx; }
.direct-arrow { flex: none; margin-left: auto; }
.capture-grid { display: flex; gap: 20rpx; margin-top: 20rpx; }
.capture-entry {
  flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; gap: 8rpx;
  padding: 24rpx 16rpx; border-radius: 28rpx;
  background: #FDF7E9; border: 3rpx solid #F1E7CE;
}
.capture-medal {
  width: 88rpx; height: 88rpx; margin-bottom: 4rpx; box-sizing: border-box; transform: rotate(-4deg);
  display: flex; align-items: center; justify-content: center;
}
.capture-medal.flip { transform: rotate(5deg) scaleX(-1); }
.capture-medal.flip .capture-ic { transform: scaleX(-1); }
.capture-ic { flex: none; }
.capture-title { font-size: 25rpx; font-weight: 800; }
.capture-note { font-size: 20rpx; color: #8B8271; text-align: center; }

/* 最近评论 */
.sec { display: flex; align-items: center; gap: 16rpx; margin: 40rpx 4rpx 20rpx; }
.sec-ic { flex: none; }
.sec-title { position: relative; }
.sec-h { font-size: 30rpx; font-weight: 900; letter-spacing: 1rpx; white-space: nowrap; }
.sec-scribble { position: absolute; left: -4rpx; bottom: -12rpx; }
.sec-p { font-size: 21rpx; color: #8B8271; padding-top: 5rpx; white-space: nowrap; }
.sec-link { margin-left: auto; font-size: 22rpx; font-weight: 700; color: #3E9B6F; white-space: nowrap; }
.recent-card {
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.recent-empty { display: flex; flex-direction: column; align-items: center; gap: 18rpx; padding: 40rpx 28rpx; }
.empty-note { font-size: 22rpx; color: #8B8271; text-align: center; }
.empty-trail { display: flex; align-items: center; gap: 10rpx; }
.trail-stop {
  width: 92rpx; height: 92rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
}
.trail-stop.flip { transform: rotate(5deg) scaleX(-1); }
.trail-stop.flip .trail-ic { transform: scaleX(-1); }
.trail-ic { flex: none; }
.trail-squig { flex: none; }
.recent-row { display: flex; align-items: center; gap: 18rpx; padding: 22rpx 24rpx; border-bottom: 2rpx solid #F5EEDF; }
.recent-row:last-child { border-bottom: 0; }
.recent-mark {
  flex: none; width: 56rpx; height: 56rpx; border-radius: 50%; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
  background: #FFF4DF; border: 2rpx solid #EADFC2;
  font-family: "Kaiti SC", "STKaiti", "PingFang SC", sans-serif; font-size: 26rpx; font-weight: 800;
}
.recent-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.recent-place { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 25rpx; font-weight: 800; }
.recent-meta { font-size: 20rpx; color: #8B8271; }
.record-status { flex: none; padding: 6rpx 14rpx; border-radius: 999rpx; font-size: 19rpx; font-weight: 700; }
.status-pending { background: #FFF4DF; color: #8A661C; }
.status-published { background: #E5F3EA; color: #2F7D58; }
.status-rejected { background: #FDEAE4; color: #C43D2D; }
.recent-arrow { flex: none; font-size: 30rpx; font-weight: 700; color: #C9BFA8; }

/* 语音 bottom-sheet */
.capture-mask { position: fixed; inset: 0; z-index: 60; background: rgba(51, 46, 38, .45); }
.voice-panel {
  position: absolute; left: 0; right: 0; bottom: 0;
  display: flex; flex-direction: column; align-items: center;
  padding: 20rpx 32rpx calc(48rpx + env(safe-area-inset-bottom));
  background: #FFFBF3; border-radius: 44rpx 44rpx 0 0;
}
.panel-handle { width: 72rpx; height: 8rpx; border-radius: 999rpx; background: #EFE5D0; }
.voice-panel-title { margin-top: 28rpx; font-size: 32rpx; font-weight: 900; }
.voice-orbit {
  width: 168rpx; height: 168rpx; margin-top: 40rpx; border-radius: 50%; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
  background: #FFF4DF; border: 4rpx solid #EADFC2;
}
.voice-orbit.recording { border-color: #3E9B6F; animation: pulse-soft 1.2s infinite alternate; }
.voice-core {
  width: 104rpx; height: 104rpx; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(180deg, #4BAB7E 0%, #3E9B6F 52%, #35855A 100%);
}
.voice-core > view { width: 24rpx; height: 24rpx; border-radius: 50%; background: #FFFFFF; }
.voice-time { margin-top: 30rpx; font-size: 26rpx; color: #8B8271; }
.voice-cancel { margin-top: 28rpx; padding: 12rpx 40rpx; font-size: 25rpx; font-weight: 700; color: #8B8271; }

/* ===== 评分表单 ===== */
.rating-page { position: fixed; inset: 0; z-index: 10; display: flex; flex-direction: column; background: linear-gradient(180deg, #FFFDF9 0%, #FFFBF3 36%, #FFF3DF 100%); }
.rating-nav {
  flex: none; display: flex; align-items: center; gap: 16rpx;
  padding: 12rpx 32rpx 16rpx; border-bottom: 3rpx solid #EFE5D0; background: #FFFBF3;
}
.mode-back { flex: none; display: flex; align-items: center; gap: 6rpx; font-size: 25rpx; font-weight: 700; }
.back-mark { font-size: 38rpx; line-height: 1; margin-top: -4rpx; }
.rating-nav-title { flex: 1; text-align: center; font-size: 32rpx; font-weight: 900; }
.switch-method { flex: none; font-size: 24rpx; font-weight: 700; color: #3E9B6F; }
.rating-scroll { flex: 1; min-height: 0; }
.rating-canvas { padding: 8rpx 32rpx calc(48rpx + env(safe-area-inset-bottom)); }

/* 字段卡 */
.field-card {
  position: relative; margin-top: 24rpx; padding: 26rpx 28rpx 30rpx;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.tape {
  position: absolute; top: -16rpx; left: 50%; width: 128rpx; height: 36rpx;
  transform: translateX(-50%) rotate(-2deg); border-radius: 6rpx;
  background: rgba(62, 155, 111, .16); border: 2rpx solid rgba(62, 155, 111, .22);
  box-sizing: border-box;
}
.field-card.field-error { border-color: #FF6D4D; }
.field-topline { display: flex; align-items: center; justify-content: space-between; }
.field-name-row { display: flex; align-items: center; gap: 10rpx; }
.field-name { font-size: 26rpx; font-weight: 800; }
.field-city { font-size: 21rpx; color: #8B8271; }
.writing-label { display: block; margin-top: 22rpx; font-size: 30rpx; font-weight: 800; }
.place-input {
  height: 92rpx; margin-top: 18rpx; padding: 0 26rpx; box-sizing: border-box;
  background: #FDF7E9; border: 3rpx solid #F1E7CE; border-radius: 22rpx;
  font-size: 28rpx; color: #332E26;
}
.writing-placeholder { color: #C9BFA8; }
.input-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 12rpx; }
.foot-hint { font-size: 20rpx; color: #8B8271; }
.foot-hint.error { color: #FF6D4D; font-weight: 700; }
.foot-counter { font-size: 19rpx; color: #C9BFA8; }

/* 建议列表 */
.place-suggestions { margin-top: 18rpx; overflow: hidden; background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 24rpx; }
.suggestion-state { padding: 22rpx 24rpx; font-size: 21rpx; color: #8B8271; }
.suggestion-row { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 24rpx; border-bottom: 2rpx solid #F5EEDF; }
.suggestion-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.suggestion-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 26rpx; font-weight: 700; }
.suggestion-meta { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 20rpx; color: #8B8271; }
.suggestion-pick { flex: none; font-size: 22rpx; font-weight: 700; color: #3E9B6F; }
.suggestion-create { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 24rpx; }
.create-ic { flex: none; }
.create-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.create-title { font-size: 24rpx; font-weight: 700; color: #2F7D58; }
.create-note { font-size: 20rpx; color: #8B8271; }

/* 已选位置 / 地图按钮 */
.selected-place-location {
  display: flex; align-items: center; gap: 16rpx; margin-top: 18rpx;
  padding: 20rpx 24rpx; border-radius: 22rpx;
  background: #E5F3EA; border: 3rpx solid #CBE5D6;
}
.selected-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.selected-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 25rpx; font-weight: 800; color: #2F7D58; }
.selected-meta { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 20rpx; color: #8B8271; }
.selected-repick { flex: none; font-size: 22rpx; font-weight: 700; color: #3E9B6F; }
.choose-location-button {
  display: flex; align-items: center; justify-content: center; gap: 10rpx;
  width: 100%; height: 80rpx; margin-top: 18rpx; padding: 0; box-sizing: border-box;
  background: #FFF4DF; border: 3rpx solid #EFE5D0; border-radius: 999rpx;
  color: #332E26; font-size: 24rpx; font-weight: 700; line-height: 1;
}
.choose-location-button::after { border: none; }
.locate-ic { flex: none; }

/* 感受选择 */
.feeling-picker {
  display: flex; align-items: center; gap: 14rpx; margin-top: 18rpx;
  min-height: 96rpx; padding: 16rpx 24rpx; box-sizing: border-box;
  background: #FDF7E9; border: 3rpx solid #F1E7CE; border-radius: 22rpx;
}
.feeling-picker.selected { background: #E5F3EA; border-color: #CBE5D6; }
.feeling-selected { flex: 1; min-width: 0; display: flex; align-items: center; gap: 14rpx; }
.feeling-placeholder { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.feeling-copy { display: flex; flex-direction: column; gap: 4rpx; min-width: 0; }
.feeling-label { font-size: 27rpx; font-weight: 800; }
.feeling-label.placeholder { color: #8B8271; }
.feeling-desc { font-size: 21rpx; color: #8B8271; }
.picker-caret { flex: none; margin-left: auto; }
.feeling-dot { flex: none; width: 22rpx; height: 22rpx; border-radius: 50%; }
.feeling-dot.hot { background: #FF6D4D; }
.feeling-dot.recommend { background: #3E9B6F; }
.feeling-dot.neutral { background: #E0A62F; }
.feeling-dot.caution { background: #8B8271; }

/* 故事折叠卡 */
.story-entry {
  margin-top: 24rpx; overflow: hidden;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.story-trigger { display: flex; align-items: center; gap: 16rpx; padding: 26rpx 28rpx; }
.story-trigger-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.story-trigger-title { font-size: 27rpx; font-weight: 800; }
.story-trigger-note { font-size: 21rpx; color: #8B8271; }
.story-toggle { flex: none; width: 52rpx; height: 52rpx; border-radius: 50%; background: #FFF4DF; color: #8B8271; font-size: 34rpx; font-weight: 700; line-height: 52rpx; text-align: center; }
.story-content { padding: 4rpx 28rpx 28rpx; border-top: 2rpx solid #F5EEDF; }
.story-entry .story-textarea,
.story-content textarea {
  width: 100%; min-height: 200rpx; margin-top: 22rpx; padding: 20rpx 24rpx; box-sizing: border-box;
  background: #FDF7E9; border: 3rpx solid #F1E7CE; border-radius: 22rpx;
  font-size: 26rpx; line-height: 1.6; color: #332E26;
}
.record-tags { margin-top: 24rpx; }
.tags-label { font-size: 21rpx; color: #8B8271; }
.tags-row { display: flex; flex-wrap: wrap; gap: 14rpx; margin-top: 14rpx; }
.tag-chip {
  padding: 10rpx 24rpx; border-radius: 999rpx;
  background: #F6EFDF; color: #6E6553; font-size: 23rpx; font-weight: 700;
}
.tag-chip.active { background: linear-gradient(180deg, #FFD05C 0%, #FFC64B 60%, #F7B93E 100%); color: #332E26; box-shadow: 0 3rpx 0 #E0A62F; }
.detail-tools { display: flex; flex-wrap: wrap; gap: 14rpx; margin-top: 24rpx; }
.tool-chip {
  padding: 10rpx 20rpx; border-radius: 999rpx; box-sizing: border-box;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; color: #6E6553; font-size: 22rpx; font-weight: 700;
}
.tool-chip.active { background: #E5F3EA; border-color: #CBE5D6; color: #2F7D58; }
.detail-fields { margin-top: 20rpx; }
.detail-field { display: flex; align-items: center; gap: 14rpx; margin-top: 14rpx; }
.detail-field-label { flex: none; font-size: 24rpx; font-weight: 700; }
.detail-field-input { flex: 1; min-width: 0; height: 76rpx; padding: 0 20rpx; box-sizing: border-box; background: #FDF7E9; border: 3rpx solid #F1E7CE; border-radius: 18rpx; font-size: 26rpx; }
.detail-field-text { flex: 1; }
.detail-field-unit { flex: none; font-size: 22rpx; color: #8B8271; }
.photo-scroll { width: 100%; margin-top: 20rpx; }
.photo-row { display: flex; gap: 14rpx; }
.photo-row image { flex: none; width: 160rpx; height: 160rpx; border-radius: 20rpx; background: #FFF4DF; }

/* 提交区 */
.submit-hint { display: block; margin-top: 26rpx; text-align: center; font-size: 22rpx; color: #FF6D4D; font-weight: 700; }
.review-hint { display: block; margin-top: 10rpx; text-align: center; font-size: 20rpx; color: #8B8271; }
.stamp-button {
  display: flex; align-items: center; justify-content: center; gap: 12rpx;
  width: 100%; height: 96rpx; margin-top: 24rpx; padding: 0; box-sizing: border-box;
  border: none; border-radius: 999rpx;
  background: #F6EFDF; color: #8B8271;
  font-size: 30rpx; font-weight: 800; line-height: 1;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.stamp-button::after { border: none; }
.stamp-button.ready { background: linear-gradient(180deg, #FFD05C 0%, #FFC64B 60%, #F7B93E 100%); color: #332E26; box-shadow: 0 5rpx 0 #E0A62F; }
.stamp-button[disabled] { opacity: .75; }
.stamp-arrow { flex: none; }

/* 感受 bottom-sheet */
.feeling-mask { position: fixed; inset: 0; z-index: 60; background: rgba(51, 46, 38, .45); }
.feeling-sheet {
  position: absolute; left: 0; right: 0; bottom: 0;
  padding: 20rpx 32rpx calc(48rpx + env(safe-area-inset-bottom));
  background: #FFFBF3; border-radius: 44rpx 44rpx 0 0;
}
.feeling-sheet-title { display: block; margin-top: 26rpx; font-size: 32rpx; font-weight: 900; }
.feeling-sheet-desc { display: block; margin-top: 8rpx; font-size: 22rpx; color: #8B8271; }
.feeling-sheet-options { margin-top: 20rpx; }
.feeling-sheet-option {
  display: flex; align-items: center; gap: 16rpx; margin-top: 14rpx;
  padding: 24rpx 22rpx; border-radius: 24rpx; box-sizing: border-box;
  border: 3rpx solid transparent; background: #FFFFFF;
}
.feeling-sheet-option.active { background: #E5F3EA; border-color: #CBE5D6; }
.feeling-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.option-check { flex: none; }
.option-arrow { flex: none; font-size: 30rpx; font-weight: 700; color: #C9BFA8; }

@keyframes pulse-soft { to { transform: scale(1.05); } }
</style>
