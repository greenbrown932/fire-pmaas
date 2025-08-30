document.addEventListener('DOMContentLoaded', function () {
  console.log('PMaaS Platform loaded');

  // Soft entrance for cards
  const fadeIn = (els) => {
    els.forEach((el, i) => {
      el.style.opacity = '0';
      el.style.transform = 'translateY(10px) scale(.98)';
      setTimeout(() => {
        el.style.transition = 'opacity .55s cubic-bezier(.29,.63,.44,.75), transform .55s cubic-bezier(.29,.63,.44,.75)';
        el.style.opacity = '1';
        el.style.transform = 'translateY(0) scale(1)';
      }, i * 90);
    });
  };

  fadeIn(Array.from(document.querySelectorAll('sl-card')));

  // Mini charts (Chart.js)
  const makeSpark = (id, data, color = 'rgb(52,211,153)') => {
    const el = document.getElementById(id);
    if (!el) return;
    new Chart(el.getContext('2d'), {
      type: 'line',
      data: {
        labels: data.map((_, i) => i + 1),
        datasets: [{
          data,
          borderColor: color,
          borderWidth: 2,
          pointRadius: 0,
          tension: 0.35,
          fill: false
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false }, tooltip: { enabled: false } },
        scales: { x: { display: false }, y: { display: false } },
        elements: { line: { capBezierPoints: true } }
      }
    });
  };

  // Fake sample sparkline data (replace with live later)
  makeSpark('sparkTotal', [3,4,4,5,4,4,4], 'rgb(94,234,212)');
  makeSpark('sparkOcc',   [1,2,2,2,2,2,2], 'rgb(74,222,128)');
  makeSpark('sparkVac',   [2,1,1,1,1,1,1], 'rgb(248,113,113)');
  makeSpark('sparkRev',   [3800,4200,4800,5000,5300,5300,5300], 'rgb(34,197,94)');

  // Keep your existing navigation and interactions...
});

