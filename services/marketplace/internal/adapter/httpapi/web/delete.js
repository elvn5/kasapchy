(() => {
  const status = document.getElementById('status');
  let challengeId = '';
  async function call(path, body, token, method = 'POST') {
    const response = await fetch('/v1' + path, {
      method, headers: {'Content-Type': 'application/json', ...(token ? {Authorization: 'Bearer ' + token} : {})},
      ...(body ? {body: JSON.stringify(body)} : {}),
    });
    if (!response.ok) throw new Error((await response.json()).error || 'Сервис недоступен.');
    return response.status === 204 ? null : response.json();
  }
  function busy(form, value) { form.querySelector('button').disabled = value; }
  document.getElementById('request').addEventListener('submit', async event => {
    event.preventDefault(); const form = event.currentTarget; busy(form, true); status.textContent = '';
    try {
      const result = await call('/auth/email/request', {email: document.getElementById('email').value, role: document.getElementById('role').value});
      challengeId = result.challengeId;
      document.getElementById('confirm').hidden = false;
      status.textContent = 'Код отправлен. Он действует 10 минут.';
    } catch (error) { status.textContent = error.message; } finally { busy(form, false); }
  });
  document.getElementById('confirm').addEventListener('submit', async event => {
    event.preventDefault(); const form = event.currentTarget; busy(form, true); status.textContent = '';
    try {
      const session = await call('/auth/email/verify', {challengeId, code: document.getElementById('code').value});
      await call('/me', null, session.token, 'DELETE');
      form.hidden = true; document.getElementById('request').hidden = true;
      status.textContent = 'Аккаунт и связанные данные удалены.';
    } catch (error) { status.textContent = error.message; } finally { busy(form, false); }
  });
})();
