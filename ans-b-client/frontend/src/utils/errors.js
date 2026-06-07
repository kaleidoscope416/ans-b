export function errorMessage(error, fallback = '请求失败，请稍后重试') {
  if (typeof error === 'string' && error.trim()) {
    return error
  }
  if (error?.message) {
    return error.message
  }
  return fallback
}
