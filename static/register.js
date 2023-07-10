function InitPurchaseDateComponent() {
	function formatDate(year, month, day) {
		return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
	}

	const now = new Date();
	const nowStr = formatDate(now.getFullYear(), now.getMonth() + 1, now.getDate());
	$("#purchase_date").val(nowStr);
	$("#purchase_date_view").html(nowStr);

	$('#purchase_date_picker').on("click", function () {
		weui.datePicker({
			start: 2022,
			end: now.getFullYear(),
			onConfirm: function (result) {
				console.log(result);
				const selectedDate = formatDate(result[0].value, result[1].value, result[2].value);
				$("#purchase_date_view").html(selectedDate);
				$("#purchase_date").val(selectedDate);
			},
			title: '购买日期'
		});
	});
}

function InitBuyFromComponent() {
	$("#buy_from_picker").on("click", function () {
		weui.picker([
			{
				label: '实体店铺',
				value: 0,
			},
			{
				label: '淘宝店铺',
				value: 1
			},
			{
				label: '微信渠道',
				value: 3
			}
		], {
			className: 'custom-classname',
			container: 'body',
			depth: 1,
			defaultValue: [1],
			onConfirm: function (result) {
				$("#buy_from").val(result[0].value);
				$("#buy_from_view").html(result[0].label);
			},
			id: 'buy_from_selector'
		})
	})
}

function InitAgreementComponent() {
	const $agreementDialog = $("#argeement_dialog");
	const $agreementDialogWrap = $("#agreement_dialog_wrap");

	function closeDialog(o) {
		const $jsDialogWrap = o.parents('.js_dialog_wrap');
		$jsDialogWrap.attr('aria-hidden', 'true').attr('aria-modal', 'false').removeAttr('tabindex');
		$jsDialogWrap.fadeOut(300);
		$jsDialogWrap.find('.js_dialog').removeClass('weui-half-screen-dialog_show');
		setTimeout(function () {
			$('#' + $jsDialogWrap.attr('ref')).trigger('focus');
		}, 300);
	}

	$('.js_dialog_wrap').on('touchmove', function (e) {
		if (!$.contains(document.getElementById('js_wrap_content'), e.target)) {
			e.preventDefault();
		}
	});

	$('.js_close').on('click', function () {
		closeDialog($(this));
	});

	$('#show_agreement').on('click', function () {
		$agreementDialogWrap.attr('aria-hidden', 'false');
		$agreementDialogWrap.attr('aria-modal', 'true');
		$agreementDialogWrap.attr('tabindex', '0');
		$agreementDialogWrap.fadeIn(200);
		$agreementDialog.addClass('weui-half-screen-dialog_show');
		setTimeout(function () {
			$agreementDialogWrap.trigger('focus');
		}, 200)
	});

	$('#agree_button').on('click', function () {
		$("#agreement_checkbox").attr("checked", true);
		closeDialog($(this));
	})

}

function InitSubmit() {
	function checkSurname() {
		const surname = $("#surname").val().trim();
		const isValid = /^[\u4e00-\u9fa5]{1,4}$/.test(surname);
		if (!isValid) {
			return "请输入中文姓氏，且不能超过 4 个字符。"
		}
		$("#surname").val(surname);
		return null
	}

	function checkForename() {
		const forename = $("#forename").val().trim();
		const isValid = /^[\u4e00-\u9fa5]{0,4}$/.test(forename);
		if (!isValid) {
			return "请输入中文名字，且不能超过 4 个字符。"
		}
		$("#forename").val(forename);
		return null
	}

	function checkPhone() {
		const phone = $("#phone").val().trim();
		const isValid = /^1\d{10}$/.test(phone);
		if (!isValid) {
			return "请输入正确的 11 位手机号。"
		}
		$("#phone").val(phone);
		return null;
	}

	function checkEmail() {
		const email = $("#email").val().trim();
		const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
		if (!isValid) {
			return "请输入正确的邮箱。"
		}
		$("#email").val(email);
		return null;
	}

	const validators = [checkSurname, checkForename, checkPhone, checkEmail];

	$('#submit_button').on('click', function (e) {
		e.preventDefault();

		for (const f of validators) {
			const result = f();
			if (result != null) {
				weui.topTips(result);
				return;
			}
		}

		if (!$("#agreement_checkbox").is(":checked")) {
			weui.alert("请阅读并同意服务条款");
			return;
		}
		$('#submit_action').click();
	})
}

InitPurchaseDateComponent();
InitBuyFromComponent();
InitAgreementComponent();
InitSubmit();