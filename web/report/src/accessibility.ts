export function activateOnKeyboard(event, activate) {
  if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); activate(); }
}
