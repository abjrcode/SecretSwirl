import { useEffect, useRef, useState } from "react"

export function Modal({
  isOpen,
  showCloseButton = false,
  onClose,
  closedByClickOutside,
  children,
}: {
  isOpen: boolean
  showCloseButton?: boolean
  onClose?: () => void
  closedByClickOutside: boolean
  children: React.ReactNode
}) {
  const [isModalOpen, setModalOpen] = useState(isOpen)
  const modalRef = useRef<HTMLDialogElement | null>(null)

  function handleModalClose() {
    if (onClose) {
      onClose()
    }
    setModalOpen(false)
  }

  function handleKeyDown(event: React.KeyboardEvent<HTMLDialogElement>) {
    if (event.key === "Escape") {
      handleModalClose()
    }
  }

  useEffect(() => {
    setModalOpen(isOpen)
  }, [isOpen])

  useEffect(() => {
    const modalElement = modalRef.current

    if (modalElement) {
      if (isModalOpen) {
        modalElement.showModal()
      } else {
        modalElement.close()
      }
    }
  }, [isModalOpen])

  return (
    <dialog
      ref={modalRef}
      onKeyDown={handleKeyDown}
      className="modal">
      <div className="modal-box">
        {children}
        {showCloseButton && (
          <div className="modal-action">
            <button
              className="btn"
              onClick={handleModalClose}>
              Close
            </button>
          </div>
        )}
      </div>
      {closedByClickOutside && (
        <div className="modal-backdrop">
          <button onClick={handleModalClose}>close</button>
        </div>
      )}
    </dialog>
  )
}
