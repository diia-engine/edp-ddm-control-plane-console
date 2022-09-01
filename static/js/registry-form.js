let renderINITemplate = function (){
    let iniTemplate = document.getElementById('ini-template').innerHTML;
    let rendered = Mustache.render(iniTemplate, {
        "CA_NAME": $("#remote-ca-name").val(),
        "CA_HOST": $("#remote-ca-host").val(),
        "CA_PORT": $("#remote-ca-port").val(),
        "KEY_SN": $("#remote-serial-number").val(),
        "KEY_HOST": $("#remote-key-host").val(),
        "KEY_ADDRESS_MASK": $("#remote-key-mask").val(),
    });

    $("#remote-ini-config").val(rendered.trim());
};

let validateEmail = function (email) {
    const re = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
};

let validateAdmins = function (adminsLine) {
    let admins = adminsLine.split(",");
    for (let i=0;i<admins.length;i++) {
        if (!validateEmail(admins[i])) {
            return false;
        }
    }

    return true;
};

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
        };
        let errorMessage = errorMessages[inputName];
        if (errorMessage) {
            input.setCustomValidity(errorMessage);
        } else {
            input.setCustomValidity('Обов’язкове поле');
        }
    }
};

let removeAllowedKeysRow = function (e){
    $(e.target).parent().remove();
    return false;
};

let keyTypeChanged = function(e){
    if (!$(e.target).is(':visible')) {
        return false;
    }

    let typeSections = $(".key-type-section");
    typeSections.find("input").prop('disabled', true);
    typeSections.hide();

    let val = $(e.target).val(),
        visibleSection = $(`.key-type-${val}`);
    visibleSection.find('input').prop('disabled', false);
    visibleSection.show();
};

$(function () {
    $("input").on('change invalid', setValidationMessages);
    $("select").on('change invalid', setValidationMessages);

    let keyDeviceType = $("#key-device-type");
    keyDeviceType.change(keyTypeChanged);
    keyDeviceType.change();

    $(`.key-type-hardware`).find('input').change(renderINITemplate);
    renderINITemplate();

    $("#allowed-keys-add").click(function () {
        let iniTemplate = document.getElementById('allowed-keys-template').innerHTML;
        $("#allowed-keys-body").append(iniTemplate);
        let allowedKeysRemoveBtn = $(".allowed-keys-remove-btn");
        allowedKeysRemoveBtn.off();
        allowedKeysRemoveBtn.click(removeAllowedKeysRow);

        $(".allowed-keys-input").on('change invalid', setValidationMessages);

        return false;
    });

    $("#officer-dns").change(function(e){
        let val = $(e.target).val();
        let officerSSL = $("#officer-ssl");

        if (val === "") {
            officerSSL.removeAttr("required");
        } else {
            officerSSL.attr("required", "required");
        }
    });

    $("#citizen-dns").change(function (e){
        let val = $(e.target).val();
        let citizenSSL = $("#citizen-ssl");

        if (val === "") {
            citizenSSL.removeAttr("required");
        } else {
            citizenSSL.attr("required", "required");
        }
    });

    $(".hide-check").change(function (e){
        let target = $(e.target),
            dataTarget = target.data("target");

        if (target.prop("checked")) {
            $(`#${dataTarget}`).show();
        } else {
            $(`#${dataTarget}`).hide();
        }
    });
});