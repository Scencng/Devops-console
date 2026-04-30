import { ElMessageBox } from 'element-plus'

const buildDangerText = (dangerPoints = []) =>
  dangerPoints
    .filter(Boolean)
    .map((item, index) => `${index + 1}. ${item}`)
    .join('\n')

export const openKafkaRiskConfirm = async ({
  title,
  resourceName,
  actionLabel,
  dangerPoints = [],
  confirmButtonText = '确认继续',
}) => {
  const message = [
    resourceName ? `对象：${resourceName}` : '',
    actionLabel ? `操作：${actionLabel}` : '',
    dangerPoints.length > 0 ? `影响：\n${buildDangerText(dangerPoints)}` : '',
  ]
    .filter(Boolean)
    .join('\n\n')

  await ElMessageBox.confirm(message.replace(/\n/g, '<br>'), title || '高风险操作确认', {
    type: 'warning',
    confirmButtonText,
    cancelButtonText: '取消',
    distinguishCancelAndClose: true,
    dangerouslyUseHTMLString: true,
  })
}

export const confirmKafkaRiskAction = async (options) => {
  try {
    await openKafkaRiskConfirm(options)
    return true
  } catch {
    return false
  }
}
