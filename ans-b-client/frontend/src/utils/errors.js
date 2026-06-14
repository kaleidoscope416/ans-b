export function errorMessage(error, fallback = '请求失败，请稍后重试') {
  if (typeof error === 'string' && error.trim()) {
    return error
  }
  if (error?.message) {
    return error.message
  }
  return fallback
}

export function isAuthExpiredError(error) {
  const message = errorMessage(error, '')
  return (
    message.includes('登录已过期') ||
    message.includes('未登录') ||
    message.includes('missing authorization token') ||
    message.includes('login session expired') ||
    message.includes('invalid login session')
  )
}
