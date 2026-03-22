import "../styles/boot-splash.css";

const bootSplashMarkup = `
  <div class="boot-splash" role="status" aria-live="polite" aria-label="Loading application" aria-busy="true">
    <div class="boot-splash__center" aria-hidden="true">
      <span class="boot-splash__mark">
        <svg class="boot-splash__logo" viewBox="0 0 24 24" role="presentation">
          <path d="M3 11h18" />
          <path d="M5 11V7a3 3 0 0 1 3-3h8a3 3 0 0 1 3 3v4" />
          <path d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0-6 0" />
          <path d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0-6 0" />
          <path d="M10 17h4" />
        </svg>
      </span>
      <span class="boot-splash__signal">
        <span class="boot-splash__signal-bar"></span>
      </span>
    </div>
  </div>
`;

function mountBootSplash(target) {
  if (!target) {
    return;
  }
  target.innerHTML = bootSplashMarkup;
}

function dismissBootSplash(target) {
  if (!target) {
    return Promise.resolve();
  }
  const splash = target.firstElementChild;
  if (!splash) {
    target.innerHTML = "";
    return Promise.resolve();
  }
  return new Promise((resolve) => {
    let settled = false;
    const finish = () => {
      if (settled) {
        return;
      }
      settled = true;
      target.innerHTML = "";
      resolve();
    };
    splash.addEventListener("transitionend", finish, { once: true });
    splash.classList.add("is-exiting");
    window.setTimeout(finish, 320);
  });
}

const BootSplash = {
  name: "BootSplash",
  template: bootSplashMarkup,
};

export { bootSplashMarkup, mountBootSplash, dismissBootSplash };

export default BootSplash;
