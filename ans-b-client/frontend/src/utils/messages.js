export function replaceMessage(messages, id, patch) {
  return messages.map((message) => (
    message.id === id
      ? { ...message, ...patch }
      : message
  ))
}
