export interface WaitingModalState { visible: boolean; challengeId?: number; receiverId?: number }
export interface ResponseModalState { visible: boolean; challengeId?: number; senderId?: number; senderName?: string; roomId?: number }

const WAITING_KEY = 'cc_waiting_modal';
const RESPONSE_KEY = 'cc_response_modal';

export function persistChallengeModals(waiting: WaitingModalState, response: ResponseModalState) {
  try {
    sessionStorage.setItem(WAITING_KEY, JSON.stringify(waiting || { visible: false }))
    sessionStorage.setItem(RESPONSE_KEY, JSON.stringify(response || { visible: false }))
  } catch {}
}

export function restoreChallengeModals(): { waiting?: WaitingModalState; response?: ResponseModalState } {
  let waiting: WaitingModalState | undefined
  let response: ResponseModalState | undefined
  try {
    const w = sessionStorage.getItem(WAITING_KEY)
    const r = sessionStorage.getItem(RESPONSE_KEY)
    if (w) {
      const obj = JSON.parse(w)
      waiting = {
        visible: !!obj.visible,
        challengeId: obj.challengeId != null ? Number(obj.challengeId) : undefined,
        receiverId: obj.receiverId != null ? Number(obj.receiverId) : undefined,
      }
    }
    if (r) {
      const obj = JSON.parse(r)
      response = {
        visible: !!obj.visible,
        challengeId: obj.challengeId != null ? Number(obj.challengeId) : undefined,
        senderId: obj.senderId != null ? Number(obj.senderId) : undefined,
        senderName: obj.senderName || undefined,
        roomId: obj.roomId != null ? Number(obj.roomId) : undefined,
      }
    }
  } catch {}
  return { waiting, response }
}
