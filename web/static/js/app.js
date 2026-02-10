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
