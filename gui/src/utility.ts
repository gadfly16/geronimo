// This is not jQuery, but a helper function to turn a html string into a HTMLElement
const _dollarRegexp = /^\s+|\s+$|(?<=\>)\s+(?=\<)/gm
export function $(html: string): HTMLElement {
  const template = document.createElement("template")
  template.innerHTML = html.replace(_dollarRegexp, "")
  const result = template.content.firstElementChild
  return result as HTMLElement
}
