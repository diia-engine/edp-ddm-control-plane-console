let setValidationMessages = function(){
    let input = $(this).get(0);
    input.setCustomValidity('');
    let inputName = $(input).attr("name");

    if (inputName === "name" && $(input).val().length < 3) {
        input.setAttribute("isvalid", "true");
        input.setCustomValidity('Будь-ласка вкажіть назву у відповідному форматі');
    }

    if (inputName === "admins" && !validateAdmins($(input).val())) {
        input.setAttribute("isvalid", "true");
        input.setCustomValidity('Будь-ласка вкажіть адміністраторів у відповідному форматі');
    }

    if (!input.validity.valid) {
        input.scrollIntoView();

        let errorMessages = {
            'name':   'Будь-ласка вкажіть назву у відповідному форматі',
            'key6': 'Будь-ласка оберіть файловий ключ',
            'sign-key-issuer': 'Будь-ласка вкажіть АЦСК, що видав ключ',
            'sign-key-pwd': 'Будь-ласка вкажіть пароль до файлового ключа',
            'ca-cert': 'Будь-ласка оберіть публічні сертифікати АЦСК',
            'ca-json': 'Будь-ласка оберіть список АЦСК',
            'admins': 'Будь-ласка вкажіть адміністраторів у відповідному форматі',
            'officer-dns': 'Невірний формат',
            'citizen-dns': 'Невірний формат',
        };
        let errorMessage = errorMessages[inputName];
        if (errorMessage) {
            input.setCustomValidity(errorMessage);
        } else {
            input.setCustomValidity('Обов’язкове поле');
        }
    }
};

$(function () {
    $("input").on('change invalid', setValidationMessages);
    $("select").on('change invalid', setValidationMessages);

    // let keyDeviceType = $("#key-device-type");
    // keyDeviceType.change(keyTypeChanged);
    // keyDeviceType.change();

    // $(`.key-type-hardware`).find('input').change(renderINITemplate);
    // renderINITemplate();

    // $("#allowed-keys-add").click(function () {
    //     let iniTemplate = document.getElementById('allowed-keys-template').innerHTML;
    //     $("#allowed-keys-body").append(iniTemplate);
    //     let allowedKeysRemoveBtn = $(".allowed-keys-remove-btn");
    //     allowedKeysRemoveBtn.off();
    //     allowedKeysRemoveBtn.click(removeAllowedKeysRow);
    //
    //     $(".allowed-keys-input").on('change invalid', setValidationMessages);
    //
    //     return false;
    // });

    // $("#officer-dns").change(function(e){
    //     let val = $(e.target).val();
    //     let officerSSL = $("#officer-ssl");
    //
    //     if (val === "") {
    //         officerSSL.removeAttr("required");
    //     } else {
    //         officerSSL.attr("required", "required");
    //     }
    // });

    // $("#citizen-dns").change(function (e){
    //     let val = $(e.target).val();
    //     let citizenSSL = $("#citizen-ssl");
    //
    //     if (val === "") {
    //         citizenSSL.removeAttr("required");
    //     } else {
    //         citizenSSL.attr("required", "required");
    //     }
    // });

    // $(".hide-check").change(function (e){
    //     let target = $(e.target),
    //         dataTarget = target.data("target");
    //
    //     if (target.prop("checked")) {
    //         $(`#${dataTarget}`).show();
    //     } else {
    //         $(`#${dataTarget}`).hide();
    //     }
    // });
});