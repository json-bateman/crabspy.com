function copyRoomUrl(el) {
  navigator.clipboard.writeText(window.location.href);
  const share = el.querySelector('.icon-share');
  const check = el.querySelector('.icon-check');
  share.classList.add('hidden');
  check.classList.remove('hidden');
  setTimeout(() => {
    share.classList.remove('hidden');
    check.classList.add('hidden');
  }, 1500);
}

function toMMSS(sec) {
  sec = Math.max(0, Number(sec) || 0);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  return String(m).padStart(2, '0') + ':' + String(s).padStart(2, '0');
};
