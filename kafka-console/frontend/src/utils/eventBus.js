/**
 * 全局事件总线
 * 用于组件间通信，特别是用于在Pod管理操作后通知其他组件刷新数据
 */

// 使用EventTarget作为事件总线的实现
class EventBus {
  constructor() {
    this.eventTarget = new EventTarget()
  }

  // 添加事件监听器
  on(eventName, callback) {
    this.eventTarget.addEventListener(eventName, callback)
  }

  // 移除事件监听器
  off(eventName, callback) {
    this.eventTarget.removeEventListener(eventName, callback)
  }

  // 触发事件
  emit(eventName, detail = {}) {
    const event = new CustomEvent(eventName, { detail })
    this.eventTarget.dispatchEvent(event)
  }
}

// 创建全局事件总线实例
const eventBus = new EventBus()

export default eventBus