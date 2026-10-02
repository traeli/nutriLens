<template>
  <view class="record-page">
    <view v-if="!recordMode" class="entry-page" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <image class="entry-art" :src="entryArt" mode="aspectFill" />

      <view class="entry-content">
        <view class="entry-heading">
          <text class="entry-kicker">饭友记 · 好好吃饭</text>
          <text class="entry-title">记录这一餐</text>
        </view>

        <view class="nutrition-entry" @tap="openNutrition">
          <view><text>02</text><text>识别餐食营养</text></view>
          <text>私人小工具 · 不参与餐厅评价</text>
          <text>进入 ›</text>
        </view>

        <view class="visit-entry">
          <view class="visit-entry-head">
            <view><text>01</text><text>到店体验</text></view>
            <text>{{ city }}</text>
          </view>

          <view class="direct-entry" @tap="startVisit('manual')">
            <view class="direct-brush"></view>
            <view><text>直接记一笔</text><text>输入餐厅，再选一个真实感受</text></view>
            <text>→</text>
            <image class="direct-photo" :src="images.noodles" mode="aspectFill" />
          </view>

          <view class="capture-grid">
            <view class="capture-entry voice-entry" @tap="openVoicePanel">
              <view class="voice-symbol"><view></view><view></view><view></view><view></view><view></view></view>
              <text>说说体验</text>
              <text>语音整理后再确认</text>
            </view>
            <view class="capture-entry receipt-entry" @tap="captureReceipt">
              <view class="receipt-symbol"><view></view><view></view><view></view></view>
              <text>拍消费单</text>
              <text>识别店名与消费信息</text>
            </view>
          </view>
        </view>

        <view class="recent-submissions">
          <view class="recent-submissions-head">
            <view><text>最近提交</text><text>提交后，可以马上在这里看到进度</text></view>
            <text @tap="openMyRecords">全部 ›</text>
          </view>
          <view v-if="recentLoading" class="recent-submit-empty">正在整理记录…</view>
          <view v-else-if="!recentRecords.length" class="recent-submit-empty">还没有提交记录，认真记下第一餐吧。</view>
          <view v-for="item in recentRecords" :key="item.id" class="recent-submit-row" @tap="openSubmittedRecord(item)">
            <view class="recent-submit-mark">{{ item.place.slice(0, 1) }}</view>
            <view class="recent-submit-copy"><text>{{ item.place }}</text><text>{{ item.date }} · {{ item.conclusion }}</text></view>
            <text class="record-status" :class="`status-${item.status}`">{{ item.statusLabel }}</text>
            <text class="recent-submit-arrow">›</text>
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

    <view v-else class="rating-page">
      <view class="rating-nav" :style="{ paddingTop: `${navLayout.contentTop}px` }">
        <view class="mode-back" @tap="backToModes"><text>‹</text><text>返回</text></view>
        <text class="rating-nav-title">记一餐</text>
        <view class="switch-method" @tap="backToModes">换个方式</view>
      </view>

      <scroll-view class="rating-scroll" scroll-y scroll-with-animation :scroll-into-view="formScrollTarget">
        <view class="rating-canvas">
          <view id="restaurantField" class="restaurant-writing" :class="{ 'field-error': placeError }">
            <view class="writing-topline">
              <view><text>01</text><text>到店记录</text></view>
              <text>{{ city }}</text>
            </view>
            <text class="writing-label">这次去了哪家店？</text>
            <input
              v-model="place"
              :focus="restaurantInputFocus"
              maxlength="160"
              placeholder="输入餐厅名称"
              placeholder-class="writing-placeholder"
              confirm-type="done"
              @input="onPlaceInput"
              @focus="onPlaceFocus"
            />
            <view class="input-foot"><text :class="{ error: placeError }">{{ placeError || placeSelectionHint }}</text><text>{{ place.length }}/160</text></view>

            <view v-if="showPlaceSuggestions" class="place-suggestions">
              <view v-if="placeSearching" class="suggestion-state">正在搜索餐厅…</view>
              <view v-for="item in placeSuggestions" :key="item.id" class="suggestion-row" @tap="choosePlaceSuggestion(item)">
                <view><text>{{ item.name }}</text><text>{{ placeSuggestionMeta(item) }}</text></view><text>选择</text>
              </view>
              <view v-if="place.trim() && !placeSearching" class="suggestion-create" @tap="choosePlaceOnMap">
                <view class="plus">⌖</view><view><text>在微信地图选择这家店</text><text>自动获取准确地址和坐标</text></view>
              </view>
            </view>

            <view v-if="selectedPlaceLocation" class="selected-place-location">
              <view><text>{{ selectedPlaceLocation.name }}</text><text>{{ selectedPlaceLocation.address }}</text></view>
              <text @tap="choosePlaceOnMap">重选</text>
            </view>
            <button v-else-if="!placeId" class="choose-location-button" @tap="choosePlaceOnMap">⌖ 从微信地图选择餐厅位置</button>
          </view>

          <view id="feelingField" class="feeling-section" :class="{ 'field-error': conclusionError }">
            <view class="section-title-row"><text class="section-number">02</text><text class="feeling-title">这一顿，感觉如何？</text></view>
            <view class="feeling-picker" :class="{ selected: selectedConclusion }" @tap="openFeelingSheet">
              <view v-if="selectedConclusion" class="feeling-selected">
                <view class="feeling-dot" :class="selectedConclusion.value"></view>
                <view><text>{{ selectedConclusion.label }}</text><text>{{ selectedConclusion.desc }}</text></view>
              </view>
              <view v-else class="feeling-placeholder"><text>选择本次感受</text><text>{{ conclusionError || '请选择最接近真实体验的一项' }}</text></view>
              <text class="picker-arrow">⌄</text>
            </view>
          </view>

          <view class="story-entry" :class="{ expanded: detailsExpanded }">
            <view class="story-trigger" @tap="detailsExpanded = !detailsExpanded">
              <view><text>再多记一点（选填）</text><text>简单记两句，或直接提交审核</text></view>
              <text>{{ detailsExpanded ? '−' : '＋' }}</text>
            </view>

            <view v-if="detailsExpanded" class="story-content">
              <textarea v-model="content" maxlength="500" :placeholder="experiencePlaceholder" />

              <view class="record-tags">
                <text>体验标签（选填）</text>
                <view><text v-for="tag in availableTags" :key="tag.code" :class="{ active: selectedTags.includes(tag.code) }" @tap="toggleTag(tag.code)">{{ tag.name }}</text></view>
              </view>

              <view class="detail-tools">
                <text :class="{ active: showCost }" @tap="showCost = !showCost">＋ 人均消费</text>
                <text :class="{ active: showWait }" @tap="showWait = !showWait">＋ 排队时间</text>
                <text :class="{ active: showDish }" @tap="showDish = !showDish">＋ 推荐菜</text>
                <text @tap="chooseExperiencePhotos">＋ 添加照片</text>
              </view>

              <view v-if="showCost || showWait || showDish" class="detail-fields">
                <view v-if="showCost"><text>人均</text><input v-model="averageCost" type="digit" placeholder="0" /><text>元</text></view>
                <view v-if="showWait"><text>排队</text><input v-model="waitMinutes" type="number" placeholder="0" /><text>分钟</text></view>
                <view v-if="showDish"><text>推荐菜</text><input v-model="recommendedDish" placeholder="写下菜名" /></view>
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
            <text>{{ submitting ? '正在提交' : '提交' }}</text><text>→</text>
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
              <view><text>{{ item.label }}</text><text>{{ item.desc }}</text></view>
              <text>{{ conclusion === item.value ? '✓' : '›' }}</text>
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

function getNavLayout() {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  const statusBarHeight = Number(info.statusBarHeight || 20)
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }
  const menuTop = menu && Number(menu.top) > 0 ? Number(menu.top) : statusBarHeight + 6
  const menuHeight = menu && Number(menu.height) > 0 ? Number(menu.height) : 32
  const menuBottom = menu && Number(menu.bottom) > menuTop ? Number(menu.bottom) : menuTop + menuHeight
  return { contentTop: Math.max(statusBarHeight, menuBottom) + 14 }
}

export default {
  data() {
    const selectedCity = getSelectedCity()
    return {
      navLayout: getNavLayout(),
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
    this.navLayout = getNavLayout()
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
        const submitted = await api.submitVisitRecord(created.record.id)
        this.resetVisitForm()
        this.recordMode = ''
        await this.loadRecentRecords()
        uni.showModal({
          title: '提交成功',
          content: '记录已经生成，当前状态为“待审核”。你可以在本页下方或“我的记录”查看进度。',
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
        const page = await api.getMyRecords({ limit: 10 })
        this.recentRecords = (page.items || []).filter(item => item.record.publish_status !== 'draft').slice(0, 3).map(this.normalizeSubmittedRecord)
      } catch (error) {
        console.warn('[record] recent submissions load failed:', error.message)
      } finally { this.recentLoading = false }
    },
    normalizeSubmittedRecord(item) {
      const status = item.record.publish_status
      const labels = { pending_review: '待审核', published: '审核通过', rejected: '审核不通过' }
      const conclusions = { hot: '夯', recommend: '推荐', neutral: '一般', caution: '谨慎' }
      return {
        id: item.record.id, status, statusLabel: labels[status] || '处理中',
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
.record-tags { margin-top: 20rpx; }
.record-tags > text { display: block; margin-bottom: 12rpx; color: #655f57; font-size: 22rpx; }
.record-tags > view { display: flex; flex-wrap: wrap; gap: 12rpx; }
.record-tags > view > text { padding: 10rpx 18rpx; border: 1rpx solid #c9c0b4; border-radius: 999rpx; color: #59544e; font-size: 21rpx; }
.record-tags > view > text.active { border-color: #365743; background: #365743; color: #fff; }
.nutrition-entry { position: relative; display: flex; align-items: center; margin: 22rpx 0; padding: 24rpx; border: 1rpx solid #d4c6b1; border-radius: 22rpx; background: rgba(255,255,255,.7); }
.nutrition-entry > view { display: flex; flex: 1; gap: 12rpx; font-weight: 700; }.nutrition-entry > text:nth-child(2) { color: #756d62; font-size: 19rpx; }.nutrition-entry > text:last-child { margin-left: 18rpx; font-weight: 700; }
.review-hint { display: block; margin: 24rpx 0; color: #5f5a52; font-size: 22rpx; text-align: center; }
.record-page {
  min-height: 100vh;
  padding-bottom: calc(108rpx + env(safe-area-inset-bottom));
  background: #F2E8D7;
}

.entry-page,
.rating-page {
  position: relative;
  min-height: calc(100vh - 108rpx - env(safe-area-inset-bottom));
  overflow: hidden;
  background: #F1E6D2;
}

.entry-art,
.rating-art {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.entry-art {
  opacity: .22;
}

.entry-content {
  position: relative;
  z-index: 1;
  padding: calc(env(safe-area-inset-top) + 40rpx) 30rpx 64rpx;
}

.entry-heading {
  display: flex;
  flex-direction: column;
  padding: 0 5rpx 30rpx;
}

.entry-kicker {
  color: #9B523E;
  font-size: 16rpx;
  font-weight: 700;
  letter-spacing: 5rpx;
}

.entry-title {
  margin-top: 7rpx;
  color: #20241F;
  font-family: "STKaiti", "KaiTi", "Songti SC", serif;
  font-size: 58rpx;
  font-weight: 700;
}

.visit-entry {
  padding: 29rpx;
  border: 2rpx solid rgba(255, 255, 255, .2);
  border-radius: 32rpx 25rpx 36rpx 27rpx;
  background: #2F3D34;
  box-shadow: 0 18rpx 45rpx rgba(39, 43, 36, .16);
  color: #FFFFFF;
}

.visit-entry-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.visit-entry-head > view {
  display: flex;
  align-items: center;
  gap: 13rpx;
}

.visit-entry-head view text:first-child {
  width: 46rpx;
  height: 46rpx;
  border: 1rpx solid rgba(255, 255, 255, .45);
  border-radius: 50%;
  font-size: 16rpx;
  line-height: 44rpx;
  text-align: center;
}

.visit-entry-head view text:last-child {
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 30rpx;
  font-weight: 700;
}

.visit-entry-head > text {
  color: #B8C8BB;
  font-size: 19rpx;
}

.direct-entry {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  margin-top: 25rpx;
  padding: 30rpx 24rpx;
  border-radius: 23rpx 18rpx 25rpx 20rpx;
  background: #F0C66E;
  color: #252B25;
}

.direct-brush {
  position: absolute;
  right: -30rpx;
  bottom: -42rpx;
  width: 230rpx;
  height: 100rpx;
  border-radius: 50%;
  background: rgba(192, 70, 45, .18);
  transform: rotate(-11deg);
}

.direct-entry > view:nth-child(2) {
  position: relative;
  display: flex;
  flex: 1;
  flex-direction: column;
}

.direct-entry view text:first-child {
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 32rpx;
  font-weight: 700;
}

.direct-entry view text:last-child {
  margin-top: 5rpx;
  color: rgba(37, 43, 37, .65);
  font-size: 18rpx;
}

.direct-entry > text {
  position: relative;
  font-size: 34rpx;
}

.capture-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 13rpx;
  margin-top: 14rpx;
}

.capture-entry {
  min-height: 190rpx;
  display: flex;
  flex-direction: column;
  padding: 23rpx;
  border: 1rpx solid rgba(255, 255, 255, .12);
  border-radius: 21rpx 18rpx 23rpx 19rpx;
  background: rgba(255, 255, 255, .075);
}

.capture-entry > text:nth-child(2) {
  margin-top: auto;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 25rpx;
  font-weight: 700;
}

.capture-entry > text:last-child {
  margin-top: 4rpx;
  color: rgba(255, 255, 255, .46);
  font-size: 16rpx;
}

.voice-symbol {
  height: 54rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
}

.voice-symbol > view {
  width: 5rpx;
  border-radius: 5rpx;
  background: #E6C573;
}

.voice-symbol > view:nth-child(1),
.voice-symbol > view:nth-child(5) { height: 17rpx; }
.voice-symbol > view:nth-child(2),
.voice-symbol > view:nth-child(4) { height: 34rpx; }
.voice-symbol > view:nth-child(3) { height: 49rpx; }

.receipt-symbol {
  width: 49rpx;
  height: 57rpx;
  padding: 12rpx 9rpx;
  border-radius: 4rpx 4rpx 11rpx 11rpx;
  background: #D67A5D;
}

.receipt-symbol view {
  height: 3rpx;
  margin-bottom: 7rpx;
  background: rgba(255, 255, 255, .8);
}

.capture-mask {
  position: fixed;
  inset: 0 0 calc(108rpx + env(safe-area-inset-bottom));
  z-index: 20;
  display: flex;
  align-items: flex-end;
  background: rgba(28, 31, 27, .5);
}

.voice-panel {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 18rpx 30rpx 44rpx;
  border-radius: 38rpx 38rpx 0 0;
  background: #F6EEDC;
}

.panel-handle {
  width: 72rpx;
  height: 7rpx;
  border-radius: 5rpx;
  background: #C9BFAF;
}

.voice-panel-title {
  margin-top: 31rpx;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 35rpx;
  font-weight: 700;
}

.voice-orbit {
  width: 190rpx;
  height: 190rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 30rpx;
  border: 2rpx solid rgba(73, 98, 83, .18);
  border-radius: 45% 55% 49% 51%;
  background: rgba(73, 98, 83, .09);
}

.voice-orbit.recording {
  animation: voice-pulse 1.1s ease-in-out infinite alternate;
}

.voice-core {
  width: 118rpx;
  height: 118rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 54% 46% 57% 43%;
  background: #C65036;
  transform: rotate(-7deg);
}

.voice-core view {
  width: 29rpx;
  height: 48rpx;
  border: 5rpx solid #FFFFFF;
  border-radius: 18rpx;
}

.voice-time {
  margin-top: 21rpx;
  color: #5F625C;
  font-size: 21rpx;
}

.voice-cancel {
  margin-top: 30rpx;
  padding: 14rpx 40rpx;
  color: #8A8378;
  font-size: 21rpx;
}

.rating-scroll {
  position: relative;
  z-index: 1;
  height: calc(100vh - 108rpx - env(safe-area-inset-bottom));
}

.rating-art {
  opacity: .92;
}

.rating-nav {
  position: relative;
  z-index: 2;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  padding: calc(env(safe-area-inset-top) + 11rpx) 27rpx 9rpx;
}

.mode-back {
  display: flex;
  align-items: center;
  justify-self: start;
  padding: 14rpx 16rpx 14rpx 0;
  color: #323A33;
}

.mode-back text:first-child {
  margin-right: 6rpx;
  font-size: 42rpx;
  line-height: 30rpx;
}

.mode-back text:last-child {
  font-size: 20rpx;
  font-weight: 650;
}

.rating-nav-title {
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 30rpx;
  font-weight: 700;
}

.city-stamp {
  justify-self: end;
  padding: 8rpx 12rpx;
  border: 2rpx solid #A4503A;
  color: #914531;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 18rpx;
  transform: rotate(3deg);
}

.rating-canvas {
  min-height: 1200rpx;
  padding: 80rpx 42rpx 420rpx;
}

.restaurant-writing {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.writing-prefix {
  color: #6C665C;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 23rpx;
}

.restaurant-writing input {
  width: 100%;
  height: 88rpx;
  margin-top: 8rpx;
  background: transparent;
  color: #252A24;
  font-family: "STKaiti", "KaiTi", "Songti SC", serif;
  font-size: 47rpx;
  font-weight: 700;
  letter-spacing: 3rpx;
  text-align: center;
}

.writing-placeholder {
  color: rgba(74, 72, 65, .38);
  font-weight: 500;
}

.brush-underline {
  position: relative;
  width: 78%;
  height: 15rpx;
  overflow: hidden;
  transform: rotate(-1deg);
}

.brush-underline::before,
.brush-underline view {
  content: '';
  position: absolute;
  left: 0;
  width: 100%;
  border-radius: 50%;
  background: #425649;
}

.brush-underline::before {
  top: 4rpx;
  height: 5rpx;
  opacity: .72;
}

.brush-underline view {
  top: 9rpx;
  left: 12%;
  width: 80%;
  height: 2rpx;
  opacity: .4;
}

.feeling-section {
  margin-top: 78rpx;
}

.feeling-title {
  display: block;
  color: #30352F;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 31rpx;
  font-weight: 700;
  text-align: center;
}

.rating-options {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 7rpx;
  margin-top: 27rpx;
}

.rating-option {
  position: relative;
  height: 118rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3E443E;
}

.paint-blob {
  position: absolute;
  width: 92rpx;
  height: 92rpx;
  border: 2rpx solid rgba(54, 59, 52, .2);
  background: rgba(247, 237, 218, .72);
  transition: transform .2s, background .2s;
}

.shape-1 .paint-blob { border-radius: 47% 53% 42% 58%; transform: rotate(-8deg); }
.shape-2 .paint-blob { border-radius: 55% 45% 57% 43%; transform: rotate(5deg); }
.shape-3 .paint-blob { border-radius: 43% 57% 51% 49%; transform: rotate(-4deg); }
.shape-4 .paint-blob { border-radius: 58% 42% 45% 55%; transform: rotate(7deg); }

.rating-option text {
  position: relative;
  z-index: 1;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 25rpx;
  font-weight: 700;
}

.rating-option.active .paint-blob {
  border-color: transparent;
  background: #405649;
  box-shadow: 0 8rpx 20rpx rgba(51, 68, 57, .2);
  transform: rotate(-5deg) scale(1.13);
}

.rating-option.hot.active .paint-blob { background: #B94B31; }
.rating-option.caution.active .paint-blob { background: #4A4842; }
.rating-option.neutral.active .paint-blob { background: #A88549; }
.rating-option.active text { color: #FFFFFF; }

.story-entry {
  margin-top: 58rpx;
  padding: 0 5rpx;
  border-top: 2rpx solid rgba(74, 72, 64, .2);
}

.story-entry.expanded {
  padding: 0 24rpx 25rpx;
  border: 1rpx solid rgba(81, 76, 65, .17);
  border-radius: 25rpx 19rpx 28rpx 21rpx;
  background: rgba(255, 248, 233, .82);
}

.story-trigger {
  min-height: 104rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.story-trigger > view {
  display: flex;
  flex-direction: column;
}

.story-trigger view text:first-child {
  color: #383C36;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 26rpx;
  font-weight: 700;
}

.story-trigger view text:last-child {
  margin-top: 5rpx;
  color: #817A6F;
  font-size: 17rpx;
}

.story-trigger > text {
  color: #9A4D38;
  font-size: 35rpx;
}

.story-content textarea {
  width: 100%;
  height: 230rpx;
  padding: 18rpx 4rpx;
  border-bottom: 3rpx solid rgba(69, 83, 72, .32);
  background: transparent;
  color: #333730;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 27rpx;
  line-height: 1.65;
}

.detail-tools {
  display: flex;
  flex-wrap: wrap;
  gap: 11rpx;
  margin-top: 22rpx;
}

.detail-tools text {
  padding: 12rpx 17rpx;
  border: 1rpx solid rgba(61, 78, 66, .25);
  border-radius: 48% 52% 46% 54%;
  color: #4C5D51;
  font-size: 18rpx;
}

.detail-tools text.active {
  background: #DDE5DB;
}

.detail-fields {
  margin-top: 20rpx;
  padding: 4rpx 18rpx;
  border-left: 5rpx solid #BE684C;
}

.detail-fields > view {
  min-height: 74rpx;
  display: flex;
  align-items: center;
  gap: 9rpx;
  border-bottom: 1rpx dashed rgba(75, 72, 64, .2);
  color: #6C665C;
  font-size: 20rpx;
}

.detail-fields input {
  min-width: 0;
  height: 64rpx;
  flex: 1;
  color: #2F342E;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 26rpx;
}

.photo-scroll {
  width: 100%;
  margin-top: 20rpx;
  white-space: nowrap;
}

.photo-row {
  display: inline-flex;
  gap: 12rpx;
}

.photo-row image {
  width: 146rpx;
  height: 146rpx;
  border: 5rpx solid #FFF9EC;
  border-radius: 12rpx 18rpx 10rpx 16rpx;
  transform: rotate(-2deg);
}

.stamp-button {
  position: relative;
  width: 142rpx;
  height: 142rpx;
  margin: 54rpx 0 0 auto;
  padding: 0;
  border: 5rpx double #A9A092;
  border-radius: 47% 53% 45% 55%;
  background: rgba(240, 230, 211, .72);
  color: #91877A;
  line-height: 132rpx;
  transform: rotate(-7deg);
}

.stamp-button::after {
  border: 0;
}

.stamp-button text {
  position: relative;
  z-index: 1;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 31rpx;
  font-weight: 750;
  letter-spacing: 3rpx;
}

.stamp-button view {
  position: absolute;
  right: 13rpx;
  bottom: 15rpx;
  left: 15rpx;
  height: 4rpx;
  border-radius: 50%;
  background: currentColor;
  opacity: .45;
}

.stamp-button.ready {
  border-color: #A94631;
  background: rgba(185, 69, 45, .93);
  color: #FFF7E7;
  box-shadow: 0 9rpx 25rpx rgba(157, 60, 40, .2);
}

.stamp-button[disabled] {
  opacity: 1;
}

@keyframes voice-pulse {
  from { transform: scale(.96) rotate(-2deg); }
  to { transform: scale(1.05) rotate(2deg); }
}

/* Refined editorial notebook theme. */
.record-page {
  background: #F4F0E7;
  color: #20231F;
}

.entry-page,
.rating-page {
  background: #F4F0E7;
}

.entry-art {
  top: auto;
  height: 62%;
  opacity: .1;
  object-position: center bottom;
}

.entry-content {
  padding-right: 32rpx;
  padding-left: 32rpx;
}

.entry-heading {
  padding-bottom: 38rpx;
}

.entry-kicker {
  color: #8B6B55;
  letter-spacing: 4rpx;
}

.entry-title {
  margin-top: 10rpx;
  color: #1F2822;
  font-family: "Songti SC", STSong, serif;
  font-size: 62rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.visit-entry {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  color: #20251F;
}

.visit-entry-head {
  padding: 0 2rpx 18rpx;
  border-bottom: 1rpx solid #D7D0C4;
}

.visit-entry-head view text:first-child {
  border-color: #9D5B46;
  color: #9D4D36;
}

.visit-entry-head view text:last-child {
  color: #222923;
  font-family: "Songti SC", STSong, serif;
  font-size: 31rpx;
}

.visit-entry-head > text {
  color: #817B72;
}

.direct-entry {
  min-height: 164rpx;
  margin-top: 18rpx;
  padding: 30rpx;
  border-radius: 22rpx;
  background: #293A31;
  box-shadow: 0 12rpx 30rpx rgba(35, 51, 42, .13);
  color: #FFFFFF;
}

.direct-brush {
  right: -45rpx;
  bottom: -62rpx;
  width: 280rpx;
  height: 140rpx;
  background: rgba(203, 125, 79, .22);
}

.direct-entry view text:first-child {
  color: #FFFFFF;
  font-family: "Songti SC", STSong, serif;
  font-size: 32rpx;
}

.direct-entry view text:last-child {
  color: rgba(255, 255, 255, .55);
}

.capture-grid {
  gap: 14rpx;
  margin-top: 14rpx;
}

.capture-entry {
  min-height: 188rpx;
  padding: 24rpx;
  border: 1rpx solid #DDD6CA;
  border-radius: 20rpx;
  background: rgba(255, 254, 250, .92);
  box-shadow: 0 7rpx 20rpx rgba(55, 49, 39, .045);
  color: #252A25;
}

.capture-entry > text:nth-child(2) {
  font-family: "Songti SC", STSong, serif;
  font-size: 25rpx;
}

.capture-entry > text:last-child {
  color: #8A847B;
}

.voice-symbol > view {
  background: #55705E;
}

.receipt-symbol {
  background: #B85C43;
}

.rating-art {
  top: auto;
  height: 47%;
  opacity: .13;
  object-position: center bottom;
}

.rating-nav {
  min-height: 94rpx;
  padding-right: 25rpx;
  padding-left: 25rpx;
  border-bottom: 1rpx solid rgba(204, 197, 185, .82);
  background: rgba(247, 244, 237, .97);
}

.mode-back {
  min-width: 120rpx;
  color: #253029;
}

.mode-back text:first-child {
  width: 38rpx;
  height: 38rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 8rpx;
  border: 1rpx solid #BDB6AA;
  border-radius: 50%;
  background: #FFFFFF;
  font-size: 32rpx;
}

.mode-back text:last-child {
  font-size: 22rpx;
}

.rating-nav-title {
  font-family: "Songti SC", STSong, serif;
  font-size: 29rpx;
}

.switch-method {
  justify-self: end;
  padding: 13rpx 0 13rpx 18rpx;
  color: #A14B35;
  font-size: 21rpx;
  font-weight: 650;
}

.rating-canvas {
  min-height: 1260rpx;
  padding: 72rpx 38rpx 370rpx;
}

.restaurant-writing {
  align-items: stretch;
  padding: 0 8rpx;
}

.writing-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.writing-meta text:first-child {
  color: #787168;
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 2rpx;
}

.writing-meta text:last-child {
  padding: 6rpx 11rpx;
  border: 1rpx solid #B85A42;
  color: #9D4833;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 17rpx;
  transform: rotate(2deg);
}

.restaurant-writing input {
  height: 104rpx;
  margin-top: 6rpx;
  color: #1F2822;
  font-family: "Songti SC", STSong, serif;
  font-size: 48rpx;
  font-weight: 700;
  letter-spacing: 1rpx;
  text-align: left;
}

.writing-placeholder {
  color: #A9A196;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 40rpx;
  font-weight: 400;
}

.brush-underline {
  width: 100%;
  height: 10rpx;
  transform: none;
}

.brush-underline::before {
  height: 3rpx;
  background: #344D3E;
  opacity: .8;
}

.brush-underline view {
  display: none;
}

.feeling-section {
  margin-top: 66rpx;
}

.feeling-title {
  color: #252C26;
  font-family: "Songti SC", STSong, serif;
  font-size: 30rpx;
  text-align: left;
}

.rating-options {
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
  margin-top: 22rpx;
}

.rating-option {
  height: 102rpx;
  justify-content: flex-start;
  gap: 17rpx;
  padding: 0 22rpx;
  border: 1rpx solid #D6CFC3;
  border-radius: 16rpx;
  background: rgba(255, 254, 250, .86);
  transition: background .18s, border-color .18s;
}

.paint-blob {
  position: relative;
  width: 34rpx;
  height: 34rpx;
  flex: none;
  border: 0;
  border-radius: 50%;
  background: #BBB4A8;
}

.shape-1 .paint-blob,
.shape-2 .paint-blob,
.shape-3 .paint-blob,
.shape-4 .paint-blob {
  border-radius: 50%;
  transform: none;
}

.rating-option.caution .paint-blob { background: #65635D; }
.rating-option.neutral .paint-blob { background: #B2945C; }
.rating-option.recommend .paint-blob { background: #54705E; }
.rating-option.hot .paint-blob { background: #B95038; }

.rating-option text {
  color: #383D38;
  font-family: "STKaiti", "KaiTi", serif;
  font-size: 27rpx;
}

.rating-option.active,
.rating-option.hot.active,
.rating-option.caution.active,
.rating-option.neutral.active {
  border-color: #334A3D;
  background: #334A3D;
  box-shadow: none;
}

.rating-option.active .paint-blob {
  width: 34rpx;
  height: 34rpx;
  border: 4rpx solid rgba(255, 255, 255, .72);
  box-shadow: none;
  transform: none;
}

.story-entry {
  margin-top: 44rpx;
  padding: 0;
  border-top: 1rpx solid #D4CDC1;
  border-bottom: 1rpx solid #D4CDC1;
}

.story-entry.expanded {
  padding: 0 24rpx 26rpx;
  border: 1rpx solid #D8D0C3;
  border-radius: 18rpx;
  background: rgba(255, 253, 247, .9);
}

.story-trigger {
  min-height: 100rpx;
}

.story-trigger view text:first-child {
  font-family: "Songti SC", STSong, serif;
  font-size: 25rpx;
}

.story-trigger view text:last-child {
  color: #8A8379;
}

.story-trigger > text {
  color: #A14B35;
}

.story-content textarea {
  border-bottom: 1rpx solid #AFA89D;
  font-family: "STKaiti", "KaiTi", serif;
}

.detail-tools text {
  border-color: #CBC3B7;
  border-radius: 14rpx;
  background: #FAF7F0;
}

.detail-tools text.active {
  background: #E4EAE3;
}

.stamp-button {
  width: 100%;
  height: 96rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 40rpx 0 0;
  padding: 0 30rpx;
  border: 0;
  border-radius: 17rpx;
  background: #D7D1C6;
  color: #918A80;
  line-height: 96rpx;
  transform: none;
}

.stamp-button text:first-child {
  font-family: "Songti SC", STSong, serif;
  font-size: 28rpx;
  letter-spacing: 1rpx;
}

.stamp-button text:last-child {
  font-size: 31rpx;
}

.stamp-button.ready {
  border: 0;
  background: #A94B35;
  color: #FFFFFF;
  box-shadow: 0 9rpx 22rpx rgba(151, 65, 44, .15);
}

.voice-panel {
  background: #F7F3EB;
}

.voice-panel-title {
  font-family: "Songti SC", STSong, serif;
}

.voice-orbit {
  border-radius: 50%;
}

.voice-core {
  border-radius: 50%;
  transform: none;
}

.record-page, .entry-page, .rating-page { background: #F3F3EF; }
.direct-entry { background: #1B1B19; }
.voice-symbol > view { background: #426A54; }
.writing-meta text:last-child, .switch-method, .story-trigger > text { border-color: #D64B32; color: #D64B32; }
.brush-underline::before { background: #426A54; }
.rating-option.recommend .paint-blob { background: #426A54; }
.rating-option.hot .paint-blob { background: #D64B32; }
.rating-option.active, .rating-option.hot.active, .rating-option.caution.active, .rating-option.neutral.active { border-color: #426A54; background: #426A54; }
.rating-option.hot.active { border-color: #D64B32; background: #D64B32; }
.stamp-button.ready { background: #D64B32; }

.record-page, .entry-page, .rating-page { background: #f5faec; }
.entry-page { min-height: 100vh; padding: calc(env(safe-area-inset-top) + 28rpx) 24rpx calc(145rpx + env(safe-area-inset-bottom)); }
.entry-art { display: none; }
.entry-content { position: static; padding: 0; }
.entry-heading { display: flex; flex-direction: column; padding: 8rpx 8rpx 22rpx; }
.entry-kicker { color: #747a73; font-size: 21rpx; font-weight: 600; letter-spacing: 0; }
.entry-title { margin-top: 28rpx; font-family: "Kaiti SC", STKaiti, cursive; font-size: 70rpx; font-weight: 900; line-height: 1; letter-spacing: -3rpx; }
.visit-entry { margin: 0; padding: 0; border: 0; background: transparent; }
.visit-entry-head { display: none; }
.direct-entry { position: relative; min-height: 300rpx; overflow: hidden; margin-top: 8rpx; padding: 30rpx; border-radius: 38rpx; background: #11120f; }
.direct-photo { position: absolute; right: -50rpx; top: -30rpx; width: 330rpx; height: 360rpx; border-radius: 48% 0 0 48%; clip-path: polygon(29% 0,100% 0,100% 100%,0 100%,26% 70%,12% 45%); }
.direct-entry > view:nth-child(2) { position: relative; z-index: 2; width: 55%; margin: 0; }
.direct-entry view text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 42rpx; font-weight: 900; }
.direct-entry view text:last-child { margin-top: 15rpx; color: rgba(255,255,255,.7); font-size: 22rpx; line-height: 1.45; }
.direct-entry > text { position: absolute; left: 32rpx; bottom: 28rpx; z-index: 3; width: 62rpx; height: 62rpx; border-radius: 50%; background: #c7ff35; color: #111; font-size: 39rpx; line-height: 58rpx; }
.direct-brush { display: none; }
.capture-grid { gap: 14rpx; margin-top: 16rpx; }
.capture-entry { min-height: 178rpx; padding: 24rpx; border: 0; border-radius: 30rpx; background: #fff; }
.capture-entry > text:nth-child(2) { font-family: "Kaiti SC", STKaiti, cursive; font-size: 31rpx; font-weight: 900; }
.capture-entry > text:last-child { color: #7c827a; font-size: 19rpx; }
.voice-symbol > view { background: #111; }
.receipt-symbol { color: #111; }
.recent-submissions { margin-top: 35rpx; }
.recent-submissions-head { display: flex; align-items: flex-end; justify-content: space-between; margin: 0 5rpx 15rpx; }
.recent-submissions-head > view { display: flex; flex-direction: column; }
.recent-submissions-head view text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 36rpx; font-weight: 900; }
.recent-submissions-head view text:last-child { margin-top: 4rpx; color: #7c8379; font-size: 18rpx; }
.recent-submissions-head > text { padding: 8rpx 0 4rpx 18rpx; color: #5b902f; font-size: 20rpx; font-weight: 800; }
.recent-submit-empty { padding: 36rpx 22rpx; border-radius: 28rpx; background: #fff; color: #747b72; text-align: center; font-size: 20rpx; }
.recent-submit-row { min-height: 112rpx; display: flex; align-items: center; gap: 15rpx; padding: 17rpx 19rpx; border-radius: 27rpx; background: #fff; }
.recent-submit-row + .recent-submit-row { margin-top: 10rpx; }
.recent-submit-mark { width: 64rpx; height: 64rpx; flex: none; border-radius: 20rpx; background: #11120f; color: #c7ff35; font-family: "Kaiti SC", STKaiti, cursive; font-size: 29rpx; font-weight: 900; line-height: 64rpx; text-align: center; }
.recent-submit-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.recent-submit-copy text:first-child { overflow: hidden; font-size: 23rpx; font-weight: 850; text-overflow: ellipsis; white-space: nowrap; }
.recent-submit-copy text:last-child { margin-top: 6rpx; color: #7e857b; font-size: 17rpx; }
.record-status { flex: none; padding: 8rpx 13rpx; border-radius: 18rpx; font-size: 17rpx; font-weight: 800; }
.status-pending_review { background: #fff0c6; color: #8a6411; }
.status-published { background: #e1f6d8; color: #387425; }
.status-rejected { background: #f9e1dc; color: #b23e2b; }
.recent-submit-arrow { color: #8a9187; font-size: 34rpx; }
.rating-page { color: #111; }
.rating-nav { background: rgba(245,250,236,.95); }
.rating-nav-title, .feeling-title, .story-trigger view text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-weight: 900; }
.stamp-button.ready { border-radius: 48rpx; background: #111; }

/* Visit form: aligned with the black, lime and soft-card product language. */
.rating-page {
  min-height: 100vh;
  background: #f5faec;
}

.rating-nav {
  min-height: 76rpx;
  padding-right: 24rpx;
  padding-bottom: 18rpx;
  padding-left: 24rpx;
  border-bottom: 0;
  background: #f5faec;
}

.mode-back {
  min-width: 112rpx;
  padding: 8rpx 0;
  color: #171914;
}

.mode-back text:first-child {
  width: 50rpx;
  height: 50rpx;
  margin-right: 9rpx;
  border: 0;
  background: #ffffff;
  font-size: 38rpx;
  line-height: 48rpx;
}

.mode-back text:last-child {
  font-size: 21rpx;
  font-weight: 750;
}

.rating-nav-title {
  font-size: 33rpx;
  line-height: 1;
}

.switch-method {
  padding: 11rpx 17rpx;
  border-radius: 24rpx;
  background: #e7ecdf;
  color: #4f574c;
  font-size: 19rpx;
  font-weight: 750;
}

.rating-scroll {
  height: calc(100vh - 108rpx - env(safe-area-inset-bottom));
}

.rating-canvas {
  min-height: 0;
  padding: 8rpx 24rpx 190rpx;
}

.restaurant-writing {
  align-items: stretch;
  padding: 29rpx;
  border-radius: 38rpx;
  background: #11120f;
  box-shadow: 0 18rpx 40rpx rgba(20, 24, 17, .14);
  color: #ffffff;
}

.writing-topline {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.writing-topline > view {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.writing-topline view text:first-child,
.section-number {
  width: 43rpx;
  height: 43rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #c7ff35;
  color: #11120f;
  font-size: 16rpx;
  font-weight: 900;
}

.writing-topline view text:last-child {
  color: rgba(255, 255, 255, .68);
  font-size: 19rpx;
  font-weight: 750;
  letter-spacing: 1rpx;
}

.writing-topline > text {
  padding: 9rpx 16rpx;
  border: 1rpx solid rgba(255, 255, 255, .2);
  border-radius: 24rpx;
  color: #c7ff35;
  font-size: 18rpx;
  font-weight: 750;
}

.writing-label {
  margin-top: 34rpx;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 36rpx;
  font-weight: 900;
}

.restaurant-writing input {
  box-sizing: border-box;
  width: 100%;
  height: 96rpx;
  margin-top: 19rpx;
  padding: 0 22rpx;
  border: 2rpx solid rgba(255, 255, 255, .12);
  border-radius: 23rpx;
  background: rgba(255, 255, 255, .08);
  color: #ffffff;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", sans-serif;
  font-size: 29rpx;
  font-weight: 750;
  letter-spacing: 0;
  text-align: left;
}

.writing-placeholder {
  color: rgba(255, 255, 255, .38);
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", sans-serif;
  font-size: 27rpx;
  font-weight: 500;
}

.input-foot {
  display: flex;
  justify-content: space-between;
  margin-top: 13rpx;
  color: rgba(255, 255, 255, .42);
  font-size: 16rpx;
}

.feeling-section {
  margin-top: 17rpx;
  padding: 28rpx;
  border-radius: 34rpx;
  background: #ffffff;
}

.section-title-row {
  display: flex;
  align-items: center;
  gap: 14rpx;
}

.feeling-title {
  color: #171914;
  font-size: 31rpx;
  text-align: left;
}

.rating-options {
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
  margin-top: 23rpx;
}

.rating-option {
  height: 92rpx;
  padding: 0 20rpx;
  border: 0;
  border-radius: 25rpx;
  background: #f0f4eb;
}

.paint-blob {
  width: 28rpx;
  height: 28rpx;
}

.rating-option text {
  color: #30352f;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", sans-serif;
  font-size: 23rpx;
  font-weight: 750;
}

.rating-option.active,
.rating-option.hot.active,
.rating-option.caution.active,
.rating-option.neutral.active {
  border: 0;
  background: #c7ff35;
  box-shadow: inset 0 0 0 2rpx #11120f;
}

.rating-option.active .paint-blob {
  width: 28rpx;
  height: 28rpx;
  border: 7rpx solid #11120f;
  background: #c7ff35;
}

.rating-option.active text {
  color: #11120f;
}

.story-entry,
.story-entry.expanded {
  margin-top: 17rpx;
  padding: 0 27rpx;
  border: 0;
  border-radius: 34rpx;
  background: #ffffff;
}

.story-entry.expanded {
  padding-bottom: 27rpx;
}

.story-trigger {
  min-height: 116rpx;
}

.story-trigger view text:first-child {
  color: #171914;
  font-size: 29rpx;
}

.story-trigger view text:last-child {
  margin-top: 7rpx;
  color: #7d8478;
  font-size: 18rpx;
}

.story-trigger > text {
  width: 52rpx;
  height: 52rpx;
  border-radius: 50%;
  background: #11120f;
  color: #c7ff35;
  font-size: 32rpx;
  line-height: 48rpx;
  text-align: center;
}

.story-content textarea {
  box-sizing: border-box;
  height: 220rpx;
  padding: 21rpx;
  border: 0;
  border-radius: 23rpx;
  background: #f1f4ed;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", sans-serif;
  font-size: 24rpx;
  line-height: 1.6;
}

.detail-tools text {
  border: 0;
  border-radius: 22rpx;
  background: #edf1e8;
  color: #485144;
}

.detail-tools text.active {
  background: #dff7a7;
  color: #26320f;
}

.detail-fields {
  border-left-color: #c7ff35;
}

.stamp-button {
  width: 100%;
  height: 98rpx;
  margin: 18rpx 0 0;
  padding: 0 31rpx;
  border-radius: 49rpx;
  background: #e0e4dc;
  color: #979d94;
  line-height: 98rpx;
}

.stamp-button text:first-child {
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", sans-serif;
  font-size: 26rpx;
  font-weight: 850;
}

.stamp-button.ready {
  border-radius: 49rpx;
  background: #11120f;
  color: #c7ff35;
  box-shadow: 0 13rpx 28rpx rgba(17, 18, 15, .16);
}

.restaurant-writing,
.feeling-section {
  transition: box-shadow .18s, transform .18s;
}

.restaurant-writing.field-error {
  box-shadow: 0 0 0 5rpx rgba(229, 82, 55, .9), 0 18rpx 40rpx rgba(20, 24, 17, .14);
}

.input-foot .error {
  color: #ff9b87;
  font-weight: 750;
}

.place-suggestions {
  overflow: hidden;
  margin-top: 17rpx;
  border: 1rpx solid rgba(255, 255, 255, .12);
  border-radius: 24rpx;
  background: #20231e;
}

.suggestion-state {
  padding: 25rpx 21rpx;
  color: rgba(255, 255, 255, .5);
  font-size: 19rpx;
}

.suggestion-row,
.suggestion-create {
  min-height: 91rpx;
  display: flex;
  align-items: center;
  gap: 15rpx;
  padding: 16rpx 19rpx;
  border-bottom: 1rpx solid rgba(255, 255, 255, .09);
}

.suggestion-row > view,
.suggestion-create > view:last-child {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
}

.suggestion-row view text:first-child,
.suggestion-create view text:first-child {
  overflow: hidden;
  color: #ffffff;
  font-size: 22rpx;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-row view text:last-child,
.suggestion-create view text:last-child {
  margin-top: 5rpx;
  overflow: hidden;
  color: rgba(255, 255, 255, .46);
  font-size: 16rpx;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-row > text {
  flex: none;
  padding: 8rpx 13rpx;
  border-radius: 18rpx;
  background: rgba(199, 255, 53, .13);
  color: #c7ff35;
  font-size: 17rpx;
  font-weight: 800;
}

.suggestion-create {
  border-bottom: 0;
  background: rgba(199, 255, 53, .07);
}

.suggestion-create .plus {
  width: 48rpx;
  height: 48rpx;
  flex: none;
  border-radius: 15rpx;
  background: #c7ff35;
  color: #11120f;
  font-size: 31rpx;
  line-height: 45rpx;
  text-align: center;
}

.choose-location-button {
  width: 100%;
  height: 82rpx;
  margin-top: 17rpx;
  border: 1rpx solid rgba(199, 255, 53, .5);
  border-radius: 22rpx;
  background: rgba(199, 255, 53, .08);
  color: #c7ff35;
  font-size: 22rpx;
  font-weight: 800;
  line-height: 80rpx;
}

.selected-place-location {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-top: 17rpx;
  padding: 18rpx 20rpx;
  border: 1rpx solid rgba(199, 255, 53, .34);
  border-radius: 22rpx;
  background: rgba(199, 255, 53, .08);
}

.selected-place-location > view {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
}

.selected-place-location view text:first-child {
  color: #fff;
  font-size: 21rpx;
  font-weight: 800;
}

.selected-place-location view text:last-child {
  margin-top: 5rpx;
  overflow: hidden;
  color: rgba(255, 255, 255, .5);
  font-size: 17rpx;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-place-location > text {
  flex: none;
  color: #c7ff35;
  font-size: 19rpx;
  font-weight: 800;
}

.feeling-section.field-error {
  box-shadow: 0 0 0 4rpx rgba(229, 82, 55, .78);
}

.feeling-picker {
  min-height: 105rpx;
  display: flex;
  align-items: center;
  gap: 15rpx;
  margin-top: 22rpx;
  padding: 18rpx 21rpx;
  border: 2rpx solid #dce2d7;
  border-radius: 25rpx;
  background: #f1f5ed;
}

.feeling-picker.selected {
  border-color: #c7ff35;
  background: #efffca;
}

.feeling-placeholder,
.feeling-selected > view:last-child {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
}

.feeling-placeholder text:first-child,
.feeling-selected view text:first-child {
  color: #20241f;
  font-size: 24rpx;
  font-weight: 850;
}

.feeling-placeholder text:last-child,
.feeling-selected view text:last-child {
  margin-top: 6rpx;
  color: #7a8277;
  font-size: 17rpx;
}

.field-error .feeling-placeholder text:last-child {
  color: #c94831;
  font-weight: 750;
}

.feeling-selected {
  min-width: 0;
  display: flex;
  flex: 1;
  align-items: center;
  gap: 16rpx;
}

.feeling-dot {
  width: 31rpx;
  height: 31rpx;
  flex: none;
  border-radius: 50%;
  background: #777a73;
}

.feeling-dot.hot { background: #e05036; }
.feeling-dot.recommend { background: #4e7b5f; }
.feeling-dot.neutral { background: #b69352; }
.feeling-dot.caution { background: #65645e; }

.picker-arrow {
  flex: none;
  color: #454b43;
  font-size: 34rpx;
}

.submit-hint {
  display: block;
  margin: 22rpx 8rpx -5rpx;
  color: #9a5d25;
  font-size: 18rpx;
  font-weight: 750;
}

.feeling-mask {
  position: fixed;
  inset: 0 0 calc(108rpx + env(safe-area-inset-bottom));
  z-index: 40;
  display: flex;
  align-items: flex-end;
  background: rgba(11, 13, 10, .56);
}

.feeling-sheet {
  width: 100%;
  box-sizing: border-box;
  padding: 18rpx 24rpx 34rpx;
  border-radius: 40rpx 40rpx 0 0;
  background: #f5faec;
}

.feeling-sheet .panel-handle {
  margin: 0 auto;
}

.feeling-sheet-title {
  display: block;
  margin-top: 28rpx;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 39rpx;
  font-weight: 900;
  text-align: center;
}

.feeling-sheet-desc {
  display: block;
  margin-top: 7rpx;
  color: #747b71;
  font-size: 19rpx;
  text-align: center;
}

.feeling-sheet-options {
  margin-top: 25rpx;
}

.feeling-sheet-option {
  min-height: 103rpx;
  display: flex;
  align-items: center;
  gap: 17rpx;
  padding: 15rpx 21rpx;
  border-radius: 25rpx;
  background: #ffffff;
}

.feeling-sheet-option + .feeling-sheet-option {
  margin-top: 10rpx;
}

.feeling-sheet-option.active {
  background: #c7ff35;
  box-shadow: inset 0 0 0 2rpx #11120f;
}

.feeling-sheet-option > view:nth-child(2) {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
}

.feeling-sheet-option view text:first-child {
  font-size: 24rpx;
  font-weight: 850;
}

.feeling-sheet-option view text:last-child {
  margin-top: 5rpx;
  color: #747b71;
  font-size: 17rpx;
}

.feeling-sheet-option > text {
  width: 43rpx;
  height: 43rpx;
  border-radius: 50%;
  background: #eef2e9;
  color: #50564e;
  font-size: 25rpx;
  line-height: 41rpx;
  text-align: center;
}

.feeling-sheet-option.active > text {
  background: #11120f;
  color: #c7ff35;
}
</style>
