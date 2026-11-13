<script>
    export let error;
</script>

<div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
    <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title">Ошибка активации аккаунта</h5>
            </div>
            <div class="modal-body">
                {error}
            </div>
        </div>
    </div>
</div>
