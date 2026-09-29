<script>
export default {
  onLaunch() {
    console.log('饭友记 Launch')
  },
}
</script>

<style>
page {
  min-height: 100%;
  background: #F5FAEC;
  color: #11120E;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif;
  font-size: 28rpx;
}

view, text, image, button, input, textarea, scroll-view {
  box-sizing: border-box;
}

button::after { border: 0; }

.page-shell {
  min-height: 100vh;
  padding: 28rpx 30rpx 56rpx;
}

.surface {
  background: #FFFFFF;
  border: 1rpx solid #E0E5DB;
  border-radius: 28rpx;
}

.primary-button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 96rpx;
  margin: 0;
  padding: 0 36rpx;
  border-radius: 18rpx;
  background: #181816;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  line-height: 1;
}

.muted { color: #87827A; }
.accent { color: #79B514; }
.positive { color: #4E8B29; }
</style>
