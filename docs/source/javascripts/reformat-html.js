$(document).ready(function () {
  // Reformat all tables as Bootstrap responsive tables
  $('table.tableblock').each(function(){
    $(this).addClass('table').wrap("<div class='table-responsive'></div>");
  });
});