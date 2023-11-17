export type ActionDataResult<R, E> = {
  success: true
  result: R
} | {
  success: false
  code: E
  error: unknown
}