const statusTextMap = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已驳回',
}

export function parseTags(value) {
  return String(value || '')
    .split(/[,，\n]/)
    .map((tag) => tag.trim())
    .filter(Boolean)
}

export function submissionStatusText(status) {
  return statusTextMap[status] || status || '未知状态'
}

export function formatDateTime(value) {
  if (!value) return '-'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'

  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
