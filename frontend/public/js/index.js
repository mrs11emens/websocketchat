document.addEventListener('DOMContentLoaded', (event) => {
    const nicknameInput = document.querySelector('.nickname');
    const chatButton = document.querySelector('button');

    chatButton.addEventListener('click', () => {
        const nickname = nicknameInput.value.trim();
        if (nickname) {
            localStorage.setItem('nickname', nickname);
            localStorage.removeItem('client_id'); // Сбросить client_id при переходе
            window.location.href = 'chat';
        } else {
            alert('Пожалуйста, введите ваш никнейм.');
        }
    });
});
