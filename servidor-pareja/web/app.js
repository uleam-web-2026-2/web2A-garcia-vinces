const pages = ['home','host-login','guest-join','dashboard','event-detail','camera'];
const state = {
  token: sessionStorage.getItem('latent.token') || '',
  user: JSON.parse(sessionStorage.getItem('latent.user') || 'null'),
  guest: JSON.parse(sessionStorage.getItem('latent.guest') || 'null'),
  event: null,
  photo: ''
};
const $ = (s, root=document) => root.querySelector(s);
const $$ = (s, root=document) => [...root.querySelectorAll(s)];
function showPage(name) {
  pages.forEach(id => $('#'+id).classList.toggle('hidden', id !== name));
  window.scrollTo({top:0,behavior:'smooth'});
  if (name === 'dashboard') loadEvents();
  if (name === 'camera') loadCamera();
}
function toast(message, kind='') {
  const el=$('#toast'); el.textContent=message; el.className=`toast ${kind}`;
  clearTimeout(toast.timer); toast.timer=setTimeout(()=>el.classList.add('hidden'),3400);
}
function fail(selector, error) { $(selector).textContent=error.message || String(error); }
async function api(path, options={}, token=state.token) {
  const headers={'Content-Type':'application/json',...(options.headers||{})};
  if (token) headers.Authorization=`Bearer ${token}`;
  let response;
  try { response=await fetch(path,{...options,headers}); }
  catch { const error=new Error('Sin conexión con el servidor. La foto puede quedar guardada para sincronizarse.'); error.networkError=true; throw error; }
  let body={}; try { body=await response.json(); } catch { if(response.status===204)return null; throw new Error('Respuesta inesperada del servidor.'); }
  if(!response.ok || body.ok===false) throw new Error(body.error?.mensaje || `Error ${response.status}`);
  return body.datos;
}
function setSession(token,user,guest=null) {
  state.token=token; state.user=user; state.guest=guest;
  sessionStorage.setItem('latent.token',token); sessionStorage.setItem('latent.user',JSON.stringify(user));
  if(guest)sessionStorage.setItem('latent.guest',JSON.stringify(guest)); else sessionStorage.removeItem('latent.guest');
}
function clearSession(){state.token='';state.user=null;state.guest=null;sessionStorage.removeItem('latent.token');sessionStorage.removeItem('latent.user');sessionStorage.removeItem('latent.guest');sessionStorage.removeItem('latent.event');}
function claim(token){try{return JSON.parse(atob(token.split('.')[1].replace(/-/g,'+').replace(/_/g,'/')))}catch{return{}}}
function esc(v){return String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));}

$$('[data-page]').forEach(button=>button.addEventListener('click',()=>showPage(button.dataset.page)));
$('#theme-toggle').addEventListener('click',()=>{document.body.classList.toggle('dark');localStorage.setItem('latent.theme',document.body.classList.contains('dark')?'dark':'light');});
if(localStorage.getItem('latent.theme')==='dark')document.body.classList.add('dark');
function showConnection(){const online=navigator.onLine;$('#connection-label').textContent=online?'CONEXIÓN LISTA':'SIN SEÑAL · FOTOS EN ESTE DISPOSITIVO';document.body.classList.toggle('offline',!online);}
window.addEventListener('online',()=>{showConnection();syncQueue();});window.addEventListener('offline',showConnection);showConnection();

$('#login-form').addEventListener('submit',async event=>{
  event.preventDefault();const form=new FormData(event.currentTarget);$('#login-error').textContent='';
  try{
    const data=await api('/login',{method:'POST',body:JSON.stringify({usuario:form.get('usuario').trim(),clave:form.get('clave')})},'');
    const payload=claim(data.token);if(payload.rol!=='anfitrion')throw new Error('Esta cuenta no es de anfitrión. Entra con el código de tu evento.');
    setSession(data.token,{nombre:payload.nombre,rol:payload.rol});$('#host-name').textContent=payload.nombre;event.currentTarget.reset();showPage('dashboard');
  }catch(error){fail('#login-error',error);}
});
$('#join-form').addEventListener('submit',async event=>{
  event.preventDefault();const form=new FormData(event.currentTarget);$('#join-error').textContent='';
  try{
    const data=await api('/eventos/unirse',{method:'POST',body:JSON.stringify({codigo_acceso:form.get('codigo_acceso').trim().toUpperCase(),nombre:form.get('nombre').trim(),pin:form.get('pin')})},'');
    const guest={id:data.invitado.id,nombre:data.invitado.nombre,evento_id:data.invitado.evento_id,fotos_tomadas:data.invitado.fotos_tomadas||0};
    const payload=claim(data.token);setSession(data.token,{nombre:payload.nombre,rol:payload.rol},guest);
    $('#guest-greeting').textContent=`¡Hola, ${guest.nombre}!`;event.currentTarget.reset();showPage('camera');
  }catch(error){fail('#join-error',error);}
});
$('#logout').addEventListener('click',()=>{clearSession();showPage('home');});
$('#guest-exit').addEventListener('click',()=>{clearSession();showPage('home');});
$('#refresh-events').addEventListener('click',loadEvents);
$('#back-dashboard').addEventListener('click',()=>showPage('dashboard'));

$('#event-form').addEventListener('submit',async event=>{
  event.preventDefault();const form=new FormData(event.currentTarget);$('#event-error').textContent='';
  try{
    await api('/eventos',{method:'POST',body:JSON.stringify({nombre:form.get('nombre').trim(),disparo_por_invitado:Number(form.get('disparo_por_invitado')),max_invitado:Number(form.get('max_invitado'))})});
    event.currentTarget.reset();toast('Tu evento ya está listo.','success');loadEvents();
  }catch(error){fail('#event-error',error);}
});
async function loadEvents(){
  if(!state.user||state.user.rol!=='anfitrion'){showPage('host-login');return;}
  const box=$('#events-list');box.innerHTML='<p class="empty">Cargando…</p>';
  try{const items=await api('/eventos');if(!items.length){box.innerHTML='<p class="empty">Todavía no tienes eventos. Crea el primero.</p>';return;}
    box.innerHTML=items.map(item=>`<div class="event-row" data-id="${Number(item.id)}"><div><p class="event-name">${esc(item.nombre)}</p><p class="event-meta">${Number(item.max_invitado)} invitados · ${Number(item.disparo_por_invitado)} fotos · ${esc(item.codigo_acceso)}</p></div><span class="pill ${esc(item.estado)}">${esc(item.estado)}</span></div>`).join('');
    $$('.event-row',box).forEach(row=>row.addEventListener('click',()=>openEvent(Number(row.dataset.id))));
  }catch(error){box.innerHTML=`<p class="empty">${esc(error.message)}</p>`;}
}
async function openEvent(id){try{state.event=await api(`/eventos/${id}`);showPage('event-detail');}catch(error){toast(error.message,'error');}}
function renderEvent(){
  const item=state.event;if(!item)return;const created=item.estado==='creado',active=item.estado==='activo';
  $('#event-content').innerHTML=`<div class="event-head"><div><p class="eyebrow">CÓDIGO PARA TUS INVITADOS</p><div class="code">${esc(item.codigo_acceso)}</div><h1>${esc(item.nombre)}</h1><span class="pill ${esc(item.estado)}">${esc(item.estado)}</span></div><div class="event-actions">${created?'<button class="btn primary" data-state="activo">Iniciar fiesta ↗</button><button class="btn quiet" id="edit-event">Editar</button><button class="btn quiet" id="delete-event">Eliminar</button>':''}${active?'<button class="btn primary" data-state="finalizado">Revelar álbum ✳</button><button class="btn quiet" data-state="cancelado">Cancelar</button>':''}</div></div><div class="stats"><div><b>${Number(item.max_invitado)}</b><small>INVITADOS MÁX.</small></div><div><b>${Number(item.disparo_por_invitado)}</b><small>FOTOS POR PERSONA</small></div><div><b>${item.estado==='finalizado'?'Revelado':'—'}</b><small>ÁLBUM</small></div></div><div class="gallery-wrap"><p class="eyebrow">FOTOS DE LA CELEBRACIÓN</p><div class="gallery" id="event-gallery"><p class="empty">Cargando fotos…</p></div></div>`;
  $$('[data-state]',$('#event-content')).forEach(button=>button.addEventListener('click',async()=>{
    const next=button.dataset.state;if(next==='cancelado'&&!confirm('¿Cancelar este evento?'))return;
    try{state.event=await api(`/eventos/${item.id}`,{method:'PATCH',body:JSON.stringify({estado:next})});renderEvent();toast(next==='finalizado'?'Álbum revelado. ✳':`Evento ${next}.`,'success');}catch(error){toast(error.message,'error');}
  }));
  $('#edit-event')?.addEventListener('click',editEvent);$('#delete-event')?.addEventListener('click',deleteEvent);loadEventPhotos(item.id);
}
async function loadEventPhotos(id){const el=$('#event-gallery');if(!el)return;try{const photos=await api(`/eventos/${id}/fotos`);renderGallery(el,photos);}catch(error){el.innerHTML=`<p class="empty">${esc(error.message)}</p>`;}}
function renderGallery(el,photos){if(!photos?.length){el.innerHTML='<p class="empty">Las fotos aparecerán aquí cuando tus invitados compartan sus momentos.</p>';return;}el.innerHTML=photos.map(photo=>`<article class="tile"><img src="${esc(photo.url_archivo)}" alt="Foto del evento" loading="lazy"><p>${photo.sincronizado?'Sincronizada':'Pendiente'} · invitado ${Number(photo.invitado_id)}</p></article>`).join('');}
async function editEvent(){const item=state.event;const nombre=prompt('Nombre del evento',item.nombre);if(nombre===null)return;const max=Number(prompt('Máximo de invitados',item.max_invitado));if(!max)return;const shots=Number(prompt('Fotos por invitado',item.disparo_por_invitado));if(!shots)return;try{state.event=await api(`/eventos/${item.id}`,{method:'PUT',body:JSON.stringify({nombre,disparo_por_invitado:shots,max_invitado:max})});renderEvent();toast('Evento actualizado.','success');}catch(error){toast(error.message,'error');}}
async function deleteEvent(){if(!confirm('Eliminar este evento vacío? Esta acción no se puede deshacer.'))return;try{await api(`/eventos/${state.event.id}`,{method:'DELETE'});toast('Evento eliminado.','success');showPage('dashboard');}catch(error){toast(error.message,'error');}}

function openQueue(){return new Promise((resolve,reject)=>{const request=indexedDB.open('latent-offline',1);request.onupgradeneeded=()=>request.result.createObjectStore('photos',{keyPath:'uuid'});request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error);});}
async function queuePhoto(photo){const db=await openQueue();await new Promise((resolve,reject)=>{const tx=db.transaction('photos','readwrite');tx.objectStore('photos').put(photo);tx.oncomplete=resolve;tx.onerror=()=>reject(tx.error);});db.close();}
async function queuedPhotos(){const db=await openQueue();const data=await new Promise((resolve,reject)=>{const tx=db.transaction('photos','readonly');const req=tx.objectStore('photos').getAll();req.onsuccess=()=>resolve(req.result);req.onerror=()=>reject(req.error);});db.close();return data;}
async function removeQueued(uuid){const db=await openQueue();await new Promise((resolve,reject)=>{const tx=db.transaction('photos','readwrite');tx.objectStore('photos').delete(uuid);tx.oncomplete=resolve;tx.onerror=()=>reject(tx.error);});db.close();}
async function syncQueue(){if(!navigator.onLine)return;let pending;try{pending=await queuedPhotos();}catch{return;}for(const photo of pending){try{await api('/fotos',{method:'POST',body:JSON.stringify({uuid:photo.uuid,url_archivo:photo.url_archivo,invitado_id:photo.invitado_id})},photo.token);await removeQueued(photo.uuid);}catch(error){if(error.networkError)break;if(error.message.includes('token')||error.message.includes('sesión'))break;await removeQueued(photo.uuid);toast('Una foto pendiente no pudo sincronizarse: '+error.message,'error');}}if(state.guest)loadCamera();}
async function loadCamera(){
  if(!state.guest||!state.token){showPage('guest-join');return;}$('#guest-greeting').textContent=`¡Hola, ${state.guest.nombre}!`;
  let eventInfo=JSON.parse(sessionStorage.getItem('latent.event')||'null'),photos=[];
  try{const id=state.guest.evento_id;const [summary,album]=await Promise.all([api(`/eventos/${id}/resumen`),api(`/eventos/${id}/fotos`)]);eventInfo=summary;photos=album;sessionStorage.setItem('latent.event',JSON.stringify(summary));$('#guest-event-name').textContent=summary.nombre;}
  catch(error){if(!eventInfo){$('#guest-event-name').textContent='Sin conexión: podrás enviar las fotos al recuperar la señal.';$('#guest-gallery').innerHTML='<p class="empty">Conéctate para ver el evento. Las fotos se guardan en este dispositivo.</p>';}}
  if(eventInfo){$('#guest-event-name').textContent=eventInfo.nombre;const remaining=Math.max(0,eventInfo.disparo_por_invitado-(state.guest.fotos_tomadas||0));$('#shots-left').textContent=eventInfo.estado==='activo'?`${remaining} ${remaining===1?'foto disponible':'fotos disponibles'}`:eventInfo.estado==='finalizado'?'El álbum ya fue revelado':'La cámara aún no está activa';$('#send-photo').disabled=!state.photo||remaining===0||eventInfo.estado!=='activo';if(eventInfo.estado==='finalizado')renderGallery($('#guest-gallery'),photos);else $('#guest-gallery').innerHTML='<p class="empty">El anfitrión revelará el álbum cuando termine la fiesta.</p>';}
  const pending=await queuedPhotos().catch(()=>[]);$('#sync-note').textContent=pending.length?`${pending.length} foto(s) pendiente(s) de sincronizar.`:'La foto se guarda en este dispositivo si se pierde la señal y se sincroniza al volver.';
}
$('#photo-input').addEventListener('change',async event=>{const file=event.target.files?.[0];if(!file)return;$('#photo-error').textContent='';if(!file.type.startsWith('image/')){$('#photo-error').textContent='Selecciona una imagen.';return;}if(file.size>12*1024*1024){$('#photo-error').textContent='El archivo original debe pesar menos de 12 MB.';return;}try{state.photo=await compressImage(file);$('#photo-preview').src=state.photo;$('#photo-preview').classList.remove('hidden');$('#camera-placeholder').classList.add('hidden');$('#send-photo').disabled=false;}catch{$('#photo-error').textContent='No se pudo preparar esta imagen.';}});
async function compressImage(file){const bitmap=await createImageBitmap(file);const scale=Math.min(1,1800/Math.max(bitmap.width,bitmap.height));const canvas=document.createElement('canvas');canvas.width=Math.round(bitmap.width*scale);canvas.height=Math.round(bitmap.height*scale);canvas.getContext('2d').drawImage(bitmap,0,0,canvas.width,canvas.height);bitmap.close?.();let quality=.82;let data=canvas.toDataURL('image/jpeg',quality);while(data.length>4*1024*1024&&quality>.4){quality-=.08;data=canvas.toDataURL('image/jpeg',quality);}if(data.length>4*1024*1024)throw new Error('foto demasiado grande');return data;}
$('#send-photo').addEventListener('click',async()=>{if(!state.photo||!state.guest)return;const button=$('#send-photo');button.disabled=true;button.textContent='Guardando…';$('#photo-error').textContent='';const item={uuid:crypto.randomUUID(),url_archivo:state.photo,invitado_id:state.guest.id,token:state.token};
  try{if(!navigator.onLine)throw Object.assign(new Error('sin red'),{networkError:true});await api('/fotos',{method:'POST',body:JSON.stringify({uuid:item.uuid,url_archivo:item.url_archivo,invitado_id:item.invitado_id})});state.guest.fotos_tomadas+=1;sessionStorage.setItem('latent.guest',JSON.stringify(state.guest));toast('Foto guardada. ¡Captura otro momento!','success');}
  catch(error){if(error.networkError){try{await queuePhoto(item);state.guest.fotos_tomadas+=1;sessionStorage.setItem('latent.guest',JSON.stringify(state.guest));toast('Sin señal: guardamos la foto para sincronizarla al volver.','success');}catch{$('#photo-error').textContent='No se pudo guardar la foto sin conexión.';}}else{$('#photo-error').textContent=error.message;}}
  state.photo='';$('#photo-input').value='';$('#photo-preview').classList.add('hidden');$('#camera-placeholder').classList.remove('hidden');button.innerHTML='Guardar mi foto <span>↗</span>';await loadCamera();});

if(state.token&&state.user?.rol==='anfitrion'){ $('#host-name').textContent=state.user.nombre;showPage('dashboard'); }
else if(state.token&&state.user?.rol==='invitado'&&state.guest){showPage('camera');syncQueue();}
else{clearSession();showPage('home');}
if('serviceWorker' in navigator)navigator.serviceWorker.register('/sw.js').catch(()=>{});
