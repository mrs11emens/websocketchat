document.addEventListener('DOMContentLoaded', () => {
    let clientId = localStorage.getItem('client_id');
    if (!clientId) {
        clientId = generateClientID();
        localStorage.setItem('client_id', clientId);
    }

    const ws = new WebSocket(`ws://192.168.10.25:8083/ws?client_id=${clientId}`);
    const chatBox = document.getElementById('chat-box');
    const messageInput = document.getElementById('message-input');
    const sendButton = document.getElementById('send-button');
    const settings = document.querySelectorAll('input[name="displayMode"]');
    const userList = document.getElementById('user-list');

    ws.onopen = () => {
        console.log('Connected to the WebSocket server');
        const nickname = localStorage.getItem('nickname');
        if (nickname) {
            ws.send(`/nick ${nickname}`);
        }
    };

    ws.onmessage = (event) => {
        const messageData = event.data;

        try {
            const data = JSON.parse(messageData);

            if (data.clients) {
                userList.innerHTML = '';
                data.clients.forEach(client => {
                    const listItem = document.createElement('li');
                    listItem.textContent = `[${client.client_id}] ${client.username}`;
                    userList.appendChild(listItem);
                });
            } else if (data.mode) {
                settings.forEach(radio => {
                    if (radio.value === data.mode) {
                        radio.checked = true;
                    }
                });
            } else if (data.messages) {
                data.messages.forEach(message => {
                    const messageElement = document.createElement('p');
                    messageElement.textContent = `[${message.sender}] ${message.content}`;
                    chatBox.appendChild(messageElement);
                });
            } else {
                const messageElement = document.createElement('p');
                messageElement.textContent = messageData;
                chatBox.appendChild(messageElement);
            }
        } catch (e) {
            console.error('Error parsing message data:', e);
            const messageElement = document.createElement('p');
            messageElement.textContent = messageData;
            chatBox.appendChild(messageElement);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    ws.onclose = (event) => {
        console.log('WebSocket connection closed:', event.reason);
    };

    sendButton.addEventListener('click', () => {
        const message = messageInput.value;
        if (message) {
            ws.send(message);
            messageInput.value = '';
        }
    });

    messageInput.addEventListener('keypress', (event) => {
        if (event.key === 'Enter') {
            sendButton.click();
        }
    });

    settings.forEach(radio => {
        radio.addEventListener('change', () => {
            ws.send('/toggle');
        });
    });

    window.addEventListener('beforeunload', () => {
        ws.close();
    });

    function generateClientID() {
        return 'client_' + Math.random().toString(36).substr(2, 9);
    }
});