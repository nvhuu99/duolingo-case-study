importScripts('https://www.gstatic.com/firebasejs/10.10.0/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/10.10.0/firebase-messaging-compat.js');

firebase.initializeApp({
    apiKey: "AIzaSyBNORe4oovYPbTtYKzmx-fp21fxS9AE7Kk",
    authDomain: "duolingo-case-study.firebaseapp.com",
    projectId: "duolingo-case-study",
    storageBucket: "duolingo-case-study.firebasestorage.app",
    messagingSenderId: "859641602216",
    appId: "1:859641602216:web:a3ccea333519773c23a29d",
    measurementId: "G-GZ62PPVDKR"
});

const messaging = firebase.messaging();

messaging.onMessage((payload) => {
    console.log("Message received:", payload);
    new Notification(payload.notification.title, {
        body: payload.notification.body,
        icon: payload.notification.icon
    });
});

messaging.onBackgroundMessage((payload) => {
    console.log('[firebase-messaging-sw.js] Received background message ', payload);
    // Customize notification here
    const notificationTitle = payload.notification.title;
    const notificationOptions = {
        body: payload.notification.body,
        // icon: '/firebase-logo.png'
    };

    self.registration.showNotification(notificationTitle, notificationOptions);
});