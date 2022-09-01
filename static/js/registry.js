$(function (){
    $( document ).tooltip();

    let registryName;
    let registryInput = $("#registry-name");
    let popupFooter = $(".popup-footer");

    let hidePopup = function (e){
        $(".popup-backdrop").hide();
        // $(".delete-popup").hide();
        $(e.target).closest(".popup-window").hide();
        registryInput.val('');
    };

    let showPopup = function (deletePopupName){
        $(".popup-backdrop").show();
        // $("#delete-popup").show();
        $(deletePopupName).show();
        popupFooter.removeClass('active');
        registryInput.val('');
    };

    $(".hide-popup").click(function (e){
        hidePopup(e);
        return false;
    });

    registryInput.val('');
    registryInput.keyup(function (e) {
        popupFooter.removeClass('active');
        if (registryName.toString() === $(e.currentTarget).val()) {
            popupFooter.addClass('active');
        }
    });

    $(".delete-registry").click(function (e){
        registryName = $(e.currentTarget).data('name');
        $("#delete-name").html(registryName);

        showPopup("#delete-popup");
    });

    $(".no-delete-registry").click(function (e){
        registryName = $(e.currentTarget).data('name');
        $("#no-delete-name").html(registryName);

        showPopup("#no-delete-popup");
    });

    $("#delete-form").submit(function () {
        return registryName.toString() === registryInput.val()
    });


});

$(document).ready(function() {
    // DataTable.datetime('DD.MM.YYYY hh:mm');

    $("#registry-table").DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
            { orderable: false, targets: 0 },
            { orderable: false, targets: 4 },
            { orderable: false, targets: 5 },
            {
                targets: 3,
                render: DataTable.render.datetime('DD.MM.YYYY h:mm'),
            },
        ],
        order: [[3, 'desc']],
        language: {
            "processing": "Зачекайте...",
            "lengthMenu": "Показати _MENU_ записів",
            "zeroRecords": "Записи відсутні.",
            "info": "Записи з _START_ по _END_ із _TOTAL_ записів",
            "infoEmpty": "Записи з 0 по 0 із 0 записів",
            "infoFiltered": "(відфільтровано з _MAX_ записів)",
            "search": "Пошук:",
            "paginate": {
                "first": "Перша",
                "previous": "Попередня",
                "next": "Наступна",
                "last": "Остання"
            },
            "aria": {
                "sortAscending": ": активувати для сортування стовпців за зростанням",
                "sortDescending": ": активувати для сортування стовпців за спаданням"
            }
        }
    });
});