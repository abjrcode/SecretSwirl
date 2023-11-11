import { useWails } from "../wails-provider/wails-context"

export function ExternalLink(props: {
  href: string
  text: string
  onClicked?: () => void
}) {
  const wails = useWails()

  function onClick(event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) {
    event.preventDefault()
    wails.runtime.BrowserOpenURL(props.href)

    if (props.onClicked) {
      props.onClicked()
    }
  }

  return (
    <a
      target="_blank"
      rel="noopener noreferrer"
      className="link-primary link-hover hover:cursor-pointer"
      onClick={onClick}>
      {props.text}
    </a>
  )
}
