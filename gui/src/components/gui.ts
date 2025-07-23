class GUI extends HTMLElement {
  constructor() {
    super()
    this.attachShadow({ mode: "open" })
  }

  connectedCallback() {
    this.render()
  }

  render() {
    this.shadowRoot!.innerHTML = `
            <style>
                :host {
                    display: grid;
                    grid-template-areas:
                        "header header"
                        "tree display";
                    grid-template-rows: auto 1fr;
                    grid-template-columns: 300px 1fr;
                    height: 100vh;
                    width: 100vw;
                    margin: 0;
                    padding: 0;
                    --primary-color: #007bff;
                    --text-color: #333;
                    --border-color: #ddd;
                    --bg-color: #f8f9fa;
                }

                app-header {
                    grid-area: header;
                    border-bottom: 1px solid var(--border-color);
                }

                app-tree {
                    grid-area: tree;
                    border-right: 1px solid var(--border-color);
                    overflow: auto;
                    padding: 10px;
                }

                app-display {
                    grid-area: display;
                    overflow: auto;
                    padding: 20px;
                }

                app-footer {
                    position: fixed;
                    bottom: 0;
                    left: 0;
                    right: 0;
                    z-index: 100;
                }
            </style>

            <app-header></app-header>
            <app-tree></app-tree>
            <app-display></app-display>
            <app-footer></app-footer>
        `
  }
}
customElements.define("g-gui", GUI)
