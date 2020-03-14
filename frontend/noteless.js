(function () {
    const width = 480;
    let height = 0;

    let video = document.querySelector('.camera-feed')
    let canvas = document.querySelector('.camera-canvas')
    const result = document.querySelector('#result')
    const processbtn = document.querySelector('#process')

    navigator.mediaDevices.getUserMedia({ video: true, audio: false })
        .then((stream) => {
            video.srcObject = stream
            video.play()
        })
        .catch((err) => {
            console.error(err)
        })

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
            mode: 'no-cors',
            body: JSON.stringify({
                Image: data
            })
        })
        .then((res) => {
            res.json().then((content) => {
                console.log(content);
                result.innerHTML = content[0];
            })
        })
    })
})()