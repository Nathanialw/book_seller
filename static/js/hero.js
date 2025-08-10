const slides = document.querySelectorAll('.slide');
const dotsContainer = document.querySelector('.dots');
let currentIndex = 0;
let slideInterval;
let slideDuration = 10000;


// Create dots
slides.forEach((_, index) => {
const dot = document.createElement('span');
dot.classList.add('dot');
if (index === 0) dot.classList.add('active');
dot.addEventListener('click', () => showSlide(index));
dotsContainer.appendChild(dot);
});

const dots = document.querySelectorAll('.dot');

function showSlide(index) {
slides[currentIndex].classList.remove('active');
dots[currentIndex].classList.remove('active');

slides[index].classList.add('active');
dots[index].classList.add('active');

currentIndex = index;
}

function nextSlide() {
    let nextIndex = (currentIndex + 1) % slides.length;
showSlide(nextIndex);
}

function startSlideShow() {
    slideInterval = setInterval(nextSlide, slideDuration);
}

startSlideShow();