// route of this page is surveys/check-in-survey.html
const numSubQuestion = 8;
const numQuestion = 6;
const mobileBaseAddr = "/mobile/home"
const apiBaseAddr = "/api/";
var surveyFailModal;
window.addEventListener('load', function () {
  // var surveyFailModalEl = document.getElementById('surveyFailModal')
  // surveyFailModalEl.addEventListener('hide.bs.modal', onModalHide)
  $('#surveyFailModal').on('hide.bs.modal', onModalHide)
  surveyFailModal = new bootstrap.Modal(document.getElementById('surveyFailModal'))
})
// surveyFailModal.addEventListe

function onCloseModal() {
  // const surveyFailModal = new bootstrap.Modal(document.getElementById('surveyFailModal'))
  surveyFailModal.hide()
}

function onModalHide() {
  console.log('modalHide')
  window.location.replace(mobileBaseAddr)

}

function submitSurvey() {
  var inputArray = createInputArray();
  var QAList = getSurveyQAList(inputArray);
  console.log(JSON.stringify(QAList));
  var instID = getQueryVariable("instID");
  var gID = getQueryVariable("memberID");
  sendSurveyToDb(instID, gID, QAList, function () {
    if (QAList.every((qa) => {
      return !qa.answer_bool
    })) {
      console.log('Survey Passed!');
      window.location.replace(mobileBaseAddr + '?proceed=true');
      return;
    }
    console.log('Survey Failed!')
      // alert("You May Not Proceed. Please Contact Your Health Consultant. Thanks For Your Survey!")
    // const surveyFailModal = new bootstrap.Modal(document.getElementById('surveyFailModal'))
    surveyFailModal.show()
      // window.location.replace(mobileBaseAddr);
  })

  

  
  }

function createInputArray() {
  var inputArray = [];
  for (i = 0; i < numSubQuestion; i++) {
    var radios = document.getElementsByName("bool-0-" + i);
    var isRadioFilled = false;
    for (j = 0; j < radios.length; j++) {
      if (radios[j].checked) {
        inputArray[i] = radios[j].value;
        isRadioFilled = true;
      }
    }
    if (!isRadioFilled) {
      alert("Please answer all survey questions" + i);
      return;
    }
  }
  for (i = numSubQuestion; i < numSubQuestion + numQuestion; i++) {
    // console.log("bool-" + (i-(numSubQuestion-1)))
    var radios = document.getElementsByName(
      "bool-" + (i - (numSubQuestion - 1))
    );
    var isRadioFilled = false;
    for (j = 0; j < radios.length; j++) {
      if (radios[j].checked) {
        inputArray[i] = radios[j].value;
        isRadioFilled = true;
      }
    }
    if (!isRadioFilled) {
      alert("Please answer all survey questions" + i);
      return;
    }
  }
  inputArray[numSubQuestion + numQuestion] = document.getElementById(
    "text-0-0"
  ).value;
  console.log(`createInput, array value - ${inputArray}`);
  return inputArray;
}

function getSurveyQAList(inputArray) {
  var QAList = [];
  for (i = 0; i < inputArray.length; i++) {
    var qid = "q" + i;
    QAList[i] = {
      question: document.getElementById(qid).textContent,
    };
    if (inputArray[i] == "0" || inputArray[i] == "1") {
      QAList[i].answer_bool = inputArray[i] == "1";
    } else {
      QAList[i].answer_text = inputArray[i];
    }
  }
  return QAList;
}

function getQueryVariable(variable) {
  var query = window.location.search.substring(1);
  var vars = query.split("&");

  // console.log('vars - ', vars);
  for (var i = 0; i < vars.length; i++) {
    var pair = vars[i].split("=");
    if (decodeURIComponent(pair[0]) == variable) {
      return decodeURIComponent(pair[1]);
    }
  }
  // console.log("Query variable %s not found", variable);
}

function sendSurveyToDb(institutionID, memberID, qaList, callback) {
  const http = new XMLHttpRequest();
  var requestBody = {
    institution_id: institutionID,
    member_id: memberID,
    qa_list: qaList,
  };
  const query = apiBaseAddr + "survey/";
  http.open("POST", query, true);
  http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
  http.onreadystatechange = function () {
    if (this.readyState === 4 && this.status === 201) {
      // console.log(this.responseText)
      console.log(this.responseText);
      if (callback) {
        callback();
      }
    } else if (this.readyState === 4) {
      alert(this.responseText);
    }
  };
  try {
    http.send(JSON.stringify(requestBody));
  } catch (e) {
    alert(e);
  }
}
