$(document).ready(function () {
  $('[data-toggle="offcanvas"]').click(function () {
    $('.sidebar').show();
    $('.row-offcanvas').toggleClass('active');
  });
});