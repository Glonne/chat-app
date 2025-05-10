let token = null;
let ws = null;
let currentRoomId = null;

// 使用相对路径
const apiBaseUrl = '/api';
const wsBaseUrl = '/api/ws/'; // WebSocket 也改为相对路径

function showMessage(msg) {
    document.getElementById('auth-message').textContent = msg;
}

function register() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    fetch(`${apiBaseUrl}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    })
    .then(response => response.json())
    .then(data => showMessage(data.message || data.error))
    .catch(err => showMessage('注册失败: ' + err));
}

function loadRooms() {
    fetch(`${apiBaseUrl}/rooms`, {
        headers: { 'Authorization': `Bearer ${token}` }
    })
    .then(response => response.json())
    .then(rooms => {
        const roomList = document.getElementById('room-list');
        roomList.innerHTML = '';
        rooms.forEach(room => {
            const li = document.createElement('li');
            li.textContent = `${room.Name} (ID: ${room.ID})`;
            li.onclick = () => joinRoom(room.ID, room.Name);
            roomList.appendChild(li);
        });
    })
    .catch(err => showMessage('加载房间失败: ' + err));
}

function createRoom() {
    const name = document.getElementById('room-name').value;
    fetch(`${apiBaseUrl}/rooms`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ name })
    })
    .then(response => response.json())
    .then(data => {
        showMessage(`房间 ${data.name} 创建成功`);
        loadRooms();
    })
    .catch(err => showMessage('创建房间失败: ' + err));
}

function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    fetch(`${apiBaseUrl}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    })
    .then(response => response.json())
    .then(data => {
        if (data.token) {
            token = data.token;
            console.log("Login token:", token);
            showMessage('登录成功');
            document.getElementById('auth-section').style.display = 'none';
            document.getElementById('chat-section').style.display = 'block';
            loadRooms();
        } else {
            showMessage(data.error);
        }
    })
    .catch(err => showMessage('登录失败: ' + err));
}

function joinRoom(roomId, roomName) {
    if (ws) ws.close();
    currentRoomId = roomId;
    document.getElementById('current-room').textContent = roomName;
    document.getElementById('room-chat').style.display = 'block';
    loadMessages(roomId);

    console.log("Joining room with token:", token);
    // 使用相对路径并动态生成 WebSocket URL
    const wsUrl = `${wsBaseUrl}${roomId}?token=${token}`;
    console.log("WebSocket URL:", wsUrl);
    ws = new WebSocket(wsUrl.replace('http://', 'ws://')); // 确保协议正确
    ws.onopen = () => {
        console.log("WebSocket opened for room:", roomId);
        showMessage(`已连接到房间 ${roomName}`);
    };
    ws.onmessage = (event) => {
        console.log("Received message:", event.data);
        const messages = document.getElementById('messages');
        messages.innerHTML += `<p>${event.data}</p>`;
        messages.scrollTop = messages.scrollHeight;
    };
    ws.onerror = (error) => console.log("WebSocket error:", error);
    ws.onclose = () => {
        console.log("WebSocket closed");
        showMessage('WebSocket 断开');
    };
}

function loadMessages(roomId) {
    fetch(`${apiBaseUrl}/rooms/${roomId}/messages`, {
        headers: { 'Authorization': `Bearer ${token}` }
    })
    .then(response => response.json())
    .then(messages => {
        const messagesDiv = document.getElementById('messages');
        messagesDiv.innerHTML = '';
        messages.forEach(msg => {
            messagesDiv.innerHTML += `<p>${msg.Content}</p>`;
        });
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    })
    .catch(err => showMessage('加载消息失败: ' + err));
}

function sendMessage() {
    const input = document.getElementById('message-input');
    const message = input.value;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        console.error("WebSocket is not open:", ws ? ws.readyState : "null");
        showMessage("WebSocket 未连接，请刷新页面");
        return;
    }
    if (message) {
        console.log("Sending message:", message);
        ws.send(message);
        input.value = '';
    }
}