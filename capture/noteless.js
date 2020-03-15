(function () {
    const width = 360;
    let height = 0;

    let video = document.querySelector('#camera-feed')
    let canvas = document.querySelector('#camera-canvas')
    const result = document.querySelector('#result')
    const processbtn = document.querySelector('#process')

    var hasTouchScreen = false;
    if ("maxTouchPoints" in navigator) {
        hasTouchScreen = navigator.maxTouchPoints > 0;
    } else if ("msMaxTouchPoints" in navigator) {
        hasTouchScreen = navigator.msMaxTouchPoints > 0;
    } else {
        var mQ = window.matchMedia && matchMedia("(pointer:coarse)");
        if (mQ && mQ.media === "(pointer:coarse)") {
            hasTouchScreen = !!mQ.matches;
        } else if ('orientation' in window) {
            hasTouchScreen = true; // deprecated, but good fallback
        } else {
            // Only as a last resort, fall back to user agent sniffing
            var UA = navigator.userAgent;
            hasTouchScreen = (
                /\b(BlackBerry|webOS|iPhone|IEMobile)\b/i.test(UA) ||
                /\b(Android|Windows Phone|iPad|iPod)\b/i.test(UA)
            );
        }
    }
    if (hasTouchScreen) { // for mobile devices we're using the rear camera
        navigator.mediaDevices.getUserMedia({ video: { facingMode: { exact: "environment" } }, audio: false })
            .then((stream) => {
                video.srcObject = stream
                video.play()
            })
            .catch((err) => {
                console.error(err)
            })
    } else { // for desktop browsers, the default front-facing camera
        navigator.mediaDevices.getUserMedia({ video: true, audio: false })
            .then((stream) => {
                video.srcObject = stream
                video.play()
            })
            .catch((err) => {
                console.error(err)
            })

    }

    video.addEventListener('canplay', () => {
        height = video.videoHeight / (video.videoWidth / width)
        video.setAttribute('width', width)
        video.setAttribute('height', height)
    })

    processbtn.addEventListener('click', () => {
        let context = canvas.getContext('2d')
        canvas.width = width
        canvas.height = height
        context.drawImage(video, 0, 0, width, height)
        let data = canvas.toDataURL('image/png')
        // remove the first chunk of "data:image/png;base64,"
        data = data.substring(22)
        // call out to Noteless HTTP API (Lambda functions)
        fetch('https://st8v3ad9y8.execute-api.eu-west-1.amazonaws.com/v1/intake', {
            method: 'post',
            body: JSON.stringify({
                Image: data
            })
        })
        .then((res) => {
            res.json().then((content) => {
                console.log(content);
                result.innerHTML = "<p>" + content.Message +"</p>\n";
            })
        })
    })
})()